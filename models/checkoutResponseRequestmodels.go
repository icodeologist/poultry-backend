package models

type CartlineInput struct {
	ProductID uint    `json:"productID"`
	Unit      string  `json:"unit"`
	Quantity  float64 `json:"quantity"`
}

type CheckOutrequest struct {
	CustomerID *uint           `json:"customerID"`
	Items      []CartlineInput `json:"items"`
}

// PayPreviousCredit - when the TenderedAmount is greater and wants to pay older credit mark this true
// else mark false and do tendered - amountdue for current order and return change if PaymentMethod == cash
//
// PayThroughCredit - when TenderedAmount == 0
// customer.Balance += amountDue - no need to record change or paymentMethod
//
// partial payement when tendered amount < amount due
// customer.Balance += amountDue - tenderedAmount
// what about order balance
type CheckOutPaymentRequest struct {
	PaymentMethod     string  `json:"paymentMethod"`
	TenderedAmount    float64 `json:"tenderedAmount"`
	ChangeGiven       float64 `json:"changeGiven"`
	PayPreviousCredit bool    `json:"payPreviousCredit"`
	PayThroughCredit  bool    `json:"payThroughCredit"`
}
