1. Check out flow
- front-end send the array with a payment method attached to backend  through post req
- backend first needs to validate the json
- okay in future we will have this request to go through with a middleware which takes a code that is unique for authentication may be jwt only
- So it decodes and converts it to go data structure
- then we use that list of items that has been listed as orders from the user and we cross check the price by fetching the actual products from db
- once we have list of products the next step is we need to calculate total
- now to see what was all the items , what was the total and how much and how the use paid we need a model that stores all of this
- My thought is to go with order 
- []products list which is one order can have multiple products
- we migth also need to know how many products of each kind this use ordered so we will have to group products with quantity
- orderedProduct model which will have product id, order id, quantity 
- now we have order which has []orderedproduct and id , totalBill, paymentMethod, Balance 
- after the products are fetched create an order model
- for each of this product check how many quantities of it user is buying then create a orderedProduct object and fill that details for that one particular product 
  - ex if product with id 1 bought 10 times the order id would be 1 then 
    oderedProduct 
    id -1
    productId - 1
    orderId - 1
    quantity - 10
  - Order would be
    id- 1
    orderredItems [orderedProduct ids] - for now [1]
    totalbill = for each product do product.unitPrice * orderedProduct.quantity += totalPrice
    payment Method = "CASH"
 return the ordder but also depedns on the payment system
2. Payment System
- So once the order is returned to front end it will have a total bill for the current order
- First check the payment method on which type of payment method user opted 
- if cash then calcuate the balance and update to the user even if the user paid full amount
- second is if the user opted to upi then generate a qr code.-> always fetch the fresh upi code from the owner bank
- during the payement we will stand by and let the user pay the money
- Once he paid we again hti backend with post request the user paid this much amount for this order 
- update the balance of the user or remind user he/she has this much balance 
- If the user says its credit then we need to think differently here
- simplest way to solve this problem is directly update the balance of the user with total amont
- and then show case the user dashboard

- So once the payment is done may be front end will have a user dashboard on 
  - total orders
  - balance 
  - some other stuff etc



why am i building this
- to maintain the credit nicely
- to have the whole inventory details
- to manage and have a reliebility of data
- to reduce the mistakes by keeping simple notebook data
- to make a system that detects
 - Product that is making most of the money
 - easy to maintain user records
 - easy to use UI
 - easy to maintain the system
