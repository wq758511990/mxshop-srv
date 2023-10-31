package model

type Category struct {
	BaseModel
	Name             string      `gorm:"type:varchar(20);not null" json:"name"`
	ParentCategory   *Category   `json:"parent"`
	SubCategory      []*Category `gorm:"foreignKey:ParentCategoryID;references:ID" json:"sub_category"`
	ParentCategoryID int32       `json:"-"`
	Level            int32       `gorm:"type:int;not null;default:1" json:"level"`
	IsTab            bool        `gorm:"default:false;not null" json:"is_tab"`
}

type Brand struct {
	BaseModel
	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"varchar(200);default: '';not null"`
}

type GoodsCategoryBrand struct {
	BaseModel
	Category   Category
	CategoryID int32 `gorm:"type:int;index:idx_category_brand,unique"`

	Brand   Brand
	BrandID int32 `gorm:"type:int;index:idx_category_brand,unique"`
}

func (GoodsCategoryBrand) TableName() string {
	return "goodscategorybrand"
}

type Banner struct {
	BaseModel
	Image string `gorm:"type:varchar(200);not null"`
	Url   string `gorm:"type:string;not null"`
	Index int32  `gorm:"type:int;not null"`
}

type Goods struct {
	BaseModel
	Category   Category
	CategoryID int32 `gorm:"type:int;index:idx_category_brand"`

	Brand   Brand
	BrandID int32 `gorm:"type:int;index:idx_category_brand"`

	Name            string   `gorm:"type:varchar(50);not null"`
	GoodsSn         string   `gorm:"type:varchar(50);not null"`
	ClickNum        int32    `gorm:"type:int;default:0;not null"`
	SoldNum         int32    `gorm:"type:int;default:0;not null"`
	FavNum          int32    `gorm:"type:int;default:0;not null"`
	MarketPrice     float32  `gorm:"not null"`
	ShopPrice       float32  `gorm:"not null"`
	GoodsBrief      string   `gorm:"type:varchar(100);not null"`
	Images          GormList `gorm:"type:varchar(1000);not null"`
	DescImages      GormList `gorm:"type:varchar(1000);not null"`
	GoodsFrontImage string   `gorm:"type:varchar(200);not null"`
	OnSale          bool     `gorm:"default:false;not null"`
	ShipFree        bool     `gorm:"default:false;not null"`
	IsNew           bool     `gorm:"default:false;not null"`
	IsHot           bool     `gorm:"default:false;not null"`
}

// 性能上考虑 不采用新建一个表 去join的方式
//type GoodsImages struct {
//	GoodsID int
//	Image   string
//}
