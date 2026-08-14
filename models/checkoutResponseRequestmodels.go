package models

type CartlineInput struct {
	ProductID uint   `json:"productID"`
	Unit      string `json:"unit"`
	Quantity  uint   `json:"quantity"`
}

type CheckOutrequest struct {
	Items []CartlineInput `json:"items"`
}
