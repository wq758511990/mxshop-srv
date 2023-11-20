package handler

import (
	"context"
	"fmt"
	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/model"
	"mxshop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

// SetInv 设置库存
func (i *InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {

	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)

	inv.Goods = req.GoodsId
	inv.Stocks = req.Num

	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	// 如果直接用first需要查询主键，这里接受的goodsId并不是， 故而需要先where
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

// Sell 出售，库存-1，需要锁，普通锁会导致所有请求都被阻塞，即使扣除的商品库存不是同一个商品，故而需要用到分布式事物锁，根据id进行事物锁处理
func (i *InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "127.0.0.1:6379",
	})
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)
	tx := global.DB.Begin()

	sellDetail := model.StockSellDetail{
		OrderSn: req.OrderSn,
		Status:  1,
	}
	var details []model.GoodsDetail
	for _, detail := range req.GoodsInfo {
		details = append(details, model.GoodsDetail{
			Goods: detail.GoodsId,
			Num:   detail.Num,
		})
		var inv model.Inventory

		// 分布式锁
		mutex := rs.NewMutex(fmt.Sprintf("goods_%s", detail.GoodsId))
		if result := global.DB.Where(&model.Inventory{Goods: detail.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		if inv.Stocks < detail.Num {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		inv.Stocks -= detail.Num
		tx.Save(&inv)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}

	}
	sellDetail.Detail = details
	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) ReBack(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还： 1：订单超时归还 2. 订单创建失败，归还之前扣减的库存 3. 手动归还
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback()
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) TrySell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "127.0.0.1:6379",
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)
	tx := global.DB.Begin()

	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew

		mutex := rs.NewMutex(fmt.Sprintf("goods_%s", goodInfo.GoodsId))
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		if inv.Stocks < goodInfo.Num {
			tx.Rollback()
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		inv.Freeze += goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (i *InventoryServer) ConfirmSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "127.0.0.1:6379",
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？
	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚之前的操作
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Stocks -= goodInfo.Num
		inv.Freeze -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update inventory set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//}else{
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
	return &emptypb.Empty{}, nil
}
func (*InventoryServer) CancelSell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//扣减库存， 本地事务 [1:10,  2:5, 3: 20]
	//数据库基本的一个应用场景：数据库事务
	//并发情况之下 可能会出现超卖 1
	client := goredislib.NewClient(&goredislib.Options{
		Addr: "127.0.0.1:6379",
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)
	rs := redsync.New(pool)

	tx := global.DB.Begin()
	//m.Lock() //获取锁 这把锁有问题吗？  假设有10w的并发， 这里并不是请求的同一件商品  这个锁就没有问题了吗？
	for _, goodInfo := range req.GoodsInfo {
		var inv model.InventoryNew
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods:goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //回滚之前的操作
		//	return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		//}

		//for {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodInfo.GoodsId))
		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis分布式锁异常")
		}
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.InvalidArgument, "没有库存信息")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //回滚之前的操作
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减， 会出现数据不一致的问题 - 锁，分布式锁
		inv.Freeze -= goodInfo.Num
		tx.Save(&inv)

		if ok, err := mutex.Unlock(); !ok || err != nil {
			return nil, status.Errorf(codes.Internal, "释放redis分布式锁异常")
		}
		//update inventory set stocks = stocks-1, version=version+1 where goods=goods and version=version
		//这种写法有瑕疵，为什么？
		//零值 对于int类型来说 默认值是0 这种会被gorm给忽略掉
		//if result := tx.Model(&model.Inventory{}).Select("Stocks", "Version").Where("goods = ? and version= ?", goodInfo.GoodsId, inv.Version).Updates(model.Inventory{Stocks: inv.Stocks, Version: inv.Version+1}); result.RowsAffected == 0 {
		//	zap.S().Info("库存扣减失败")
		//}else{
		//	break
		//}
		//}
		//tx.Save(&inv)
	}
	tx.Commit() // 需要自己手动提交操作
	//m.Unlock() //释放锁
	return &emptypb.Empty{}, nil
}
