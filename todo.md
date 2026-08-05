[] - pos system
  - workers searches the items
  - items will be added or removed in the real time
  - when adding or removing the prices will be updated


=> Create a transaction flow
 - Start with ID  a unique transaction 
 - Then search items
 - That item along with its prices for that day will be added to the stack
 - Add to some list or stack - recent added will always be at the top
Stage 1 — Database: Products table

Design and create the table that holds your catalog.

Columns: product_id (unique), name, price, unit (e.g. "per kg" vs "per piece" — matters for chicken/mutton vs packaged goods), tax_rate, barcode (nullable — loose items won't have one), stock_quantity (optional, if you want inventory tracking)
Build a simple admin page: add product, edit price, delete product. This is your first working piece — test it fully before moving on.
Stage 2 — Front-end cart (in-memory, no backend yet)

Build the cashier screen as a self-contained UI, backed by nothing but a JS array.

Search/list products on screen (pull from Stage 1's data)
"Add to cart" → pushes a line into the cart array: { productId, name, price, qty }
Delete button per line → removes that line from the array by its ID
Quantity field per line → editable, recalculates that line's subtotal
Running total → recomputed from the whole array every time it changes (sum of price × qty across all lines), not tracked as a separate incrementing variable
At this stage, get this fully working and correct with fake/hardcoded product data if needed. Confirm add, delete, quantity-edit, and total all behave right before touching a database.
Stage 3 — Weight/manual entry for non-barcoded items

Extend Stage 2's "add to cart" so it handles chicken, mutton, milk, etc.

If a product's unit is "per kg" → show a weight input instead of a plain "add" click; line subtotal = price_per_kg × weight_entered
Add a text/search-based "add by name" option alongside barcode entry, since these items have no code to scan
(Barcode entry itself, when you get to it, is just a text input that adds-on-Enter — same code path as manual add, just a different trigger)
Stage 4 — Database: Orders and Order Items tables

Now design where a finalized sale gets permanently stored.

orders table: order_id, timestamp, total_amount, payment_method, status (paid/refunded/etc.)
order_items table: order_item_id, order_id (links to above), product_id (links to Stage 1), quantity, price_at_sale (the price at the time of sale, copied in — never referenced live from products, so future price changes don't rewrite history)
Stage 5 — Checkout: connect front end to backend

This is the one moment the cart array talks to the database.

Cashier hits "Finalize/Checkout"
Front end sends the whole cart array in one API call to the backend
Backend: creates one row in orders, then loops through the cart and inserts one row per item into order_items
(Optional but recommended) Backend re-checks each product's current price against what the front end sent, so a stale or tampered price can't slip through
Backend returns success + the new order_id
Front end clears the cart array, ready for the next customer
Stage 6 — Payment method

Add the actual payment step, right before or as part of Stage 5's checkout.

UI: buttons/options for cash, card, UPI, credit (whatever you support)
If cash/card/UPI → mark status = paid immediately once confirmed
If credit (customer owes you, pays later) → mark status = pending or credit, and you'll want a way to later mark it paid when they settle up
This choice just gets saved as payment_method and status on the orders row from Stage 4 — no separate system needed at this scale
Stage 7 — Invoice generation

Once an order is saved (Stage 5), generate a receipt from it.

Pull the order_items rows for that order_id, join with products for names, format as a printable/emailable invoice
This is now trivial because Stage 4's design already stores exactly what you need for a legal record — item, quantity, price-at-the-time, total
