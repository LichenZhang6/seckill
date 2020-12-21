package datamodels

type Product struct {
	ID           int64  `json:"id" sql:"id" web:"id"`
	ProductName  string `json:"ProductName" sql:"productName" web:"productName"`
	ProductNum   int64  `json:"ProductNum" sql:"productNum" web:"productNum"`
	ProductImage string `json:"ProductImage" sql:"productImage" web:"productImage"`
	ProductURL   string `json:"ProductUrl" sql:"productUrl" web:"productUrl"`
}
