**Yahoo stock simulator in GO Language**

I have built a virtual stock trading system for whoever wants to learn how to invest in stocks.
The system uses real-time pricing via Yahoo finance API and supports USD currency only. 
The system will have 2 components: Client and Server.
Client: The JSON-RPC client will take command line input and send requests to the server.
Server: The trading engine will have JSON-RPC interface for the below features.

*Feature #1:* 
Buying stock: 
A)	Request
•	User requests the server to buy stocks by giving the stock symbol and the percent of the budget he wants to invest in that stock.
•	He/She also inputs the overall budget for this transaction.

                “stockSymbolAndPercentage”: string (E.g. “GOOG: 50%, YHOO: 50 %”)
                “budget”: float32
B)	Response
•	Server then allocates the stock to the user after performing some validations. 
•	It return a tradeID, the unvested (leftover) amount and the stock details that include: Stock symbol, no of stocks purchased for that stock and the price at which the stock was bought. 
     “tradeId”: number
     “stocks”: string (E.g. “GOOG: 100:$500.25”, “YHOO: 200:$31.40”)
                 “unvestedAmount”: float32

*Feature #2:* 
Check portfolio (Loss/Gain): 
A)	Request
•	User can check the portfolio to see if the current price of the stocks and to figure out if the previous transaction were a profit or loss.
•	User only inputs the tradeId that was previously allocated to earlier transaction(s).
             “tradeId”: number

B)	Response
•	User gets the details of the stocks that were purchased w.r.t the tradeId as provided. 
•	The details also include the (+) Profit or (-) Loss symbol that compares the current price of the stocks to earlier one along with the unvested amount.
“stocks”: string (E.g. “GOOG:100:+$520.25”, “YHOO:200:-$30.40”)
“currentMarketValue” : float32
    “unvestedAmount”: float32
