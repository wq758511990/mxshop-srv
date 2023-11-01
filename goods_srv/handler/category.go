package handler

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"
)

// GetAllCategorysList 商品分类
func (s *GoodsServer) GetAllCategorysList(ctx context.Context, req *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var categories []model.Category
	global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categories)
	b, _ := json.Marshal(&categories)
	return &proto.CategoryListResponse{
		JsonData: string(b),
	}, nil
}

// GetSubCategory 获取子分类
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	categoryListResponse := proto.SubCategoryListResponse{}

	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	categoryListResponse.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		Level:          category.Level,
		IsTab:          category.IsTab,
		ParentCategory: category.ParentCategoryID,
	}

	var subCategories []model.Category
	var subCategoryResponse []*proto.CategoryInfoResponse
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Find(&subCategories)
	for _, subCategory := range subCategories {
		subCategoryResponse = append(subCategoryResponse, &proto.CategoryInfoResponse{
			Id:             subCategory.ID,
			Name:           subCategory.Name,
			Level:          subCategory.Level,
			IsTab:          subCategory.IsTab,
			ParentCategory: subCategory.ParentCategoryID,
		})
	}
	categoryListResponse.SubCategorys = subCategoryResponse
	return &categoryListResponse, nil
}

func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	var category model.Category
	if req.ParentCategory > 0 {
		if result := global.DB.First(&category, req.ParentCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "parentId不存在")
		}
	}
	categoryForm := model.Category{
		IsTab: req.IsTab,
		Name:  req.Name,
		Level: req.Level,
	}
	if req.ParentCategory > 0 {
		categoryForm.ParentCategoryID = req.ParentCategory
	}
	tx := global.DB.Begin()
	result := tx.Save(&categoryForm)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	tx.Commit()
	return &proto.CategoryInfoResponse{
		Id: categoryForm.ID,
	}, nil
}

func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "删除分类不存在")
	}
	tx := global.DB.Begin()
	// TODO: 待补充
	result := tx.Delete(&model.Category{}, req.Id)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	tx.Commit()
	var subCategories []model.Category
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Find(&subCategories)
	// 递归删除
	for _, child := range subCategories {
		s.DeleteCategory(ctx, &proto.DeleteCategoryRequest{
			Id: child.ID,
		})
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "分类不存在")
	}
	category.ID = req.Id
	category.Name = req.Name
	category.Level = req.Level
	category.IsTab = req.IsTab
	category.ParentCategoryID = req.ParentCategory

	tx := global.DB.Begin()
	result := tx.Save(&category)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}
