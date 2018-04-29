package wallet

import (
	"net/http"

	"encoding/json"
	"fmt"

	"github.com/Encrypt-S/kauri-api/app/api"
	"github.com/gorilla/mux"
	"github.com/Encrypt-S/kauri-api/app/daemon/daemonrpc"
	"github.com/Encrypt-S/kauri-api/app/conf"
)

// InitTransactionHandlers sets up handlers for transaction-related rpc commands
func InitTransactionHandlers(r *mux.Router, prefix string) {

	namespace := "transactions"

	// setup endpoint to be used for receiving txids for all supplied addresses
	getTxIdsPath := api.RouteBuilder(prefix, namespace, "v1", "getaddresstxids")
	api.OpenRouteHandler(getTxIdsPath, r, getData("txids"))

	// setup endpoint to be used for receiving raw transaction datat for all supplied addresses
	getRawTransactionsPath := api.RouteBuilder(prefix, namespace, "v1", "getrawtransactions")
	api.OpenRouteHandler(getRawTransactionsPath, r, getData("raw"))

}

type Response struct {
	Results []Result `json:"results"`
}

type Result struct {
	Currency  string `json:"currency"`
	Addresses []Address `json:"addresses"`
}

type Address struct {
	Address      string `json:"address"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
		Txid    string `json:"txid"`
		Rawtx   string `json:"rawtx"`
		Verbose string `json:"verbose"`
}


// IncomingTransactionsArray describes the Transactions array
type AddressesReq struct {
	Addresses []AddressReqItem `json:"transactions"`
}


// IncomingTransactions describes the incoming Transactions array
type AddressReqItem struct {
	Currency  string   `json:"currency"`
	Addresses []string `json:"addresses"`
}

// OutgoingTransactionsArray describes the parsed transactions data array
//type OutgoingTransactionsArray struct {
//	Transactions []OutgoingTransactions `json:"transactions"`
//}

// OutgoingTransactions describes the outgoing response
//type OutgoingTransactions struct {
//	Currency          string                  `json:"currency"`
//	OutgoingAddresses []OutgoingAddressObject `json:"addressobject"`
//}

// OutgoingAddressObject contains address and array of txids
//type OutgoingAddressObject struct {
//	Address            string   `json:"address"`
//	OutgoingTxIdsArray []string `json:"txids"`
//}

// RPCGetAddressTxIDParams describes addresses array params for getaddresstxids call
type RPCGetAddressTxIDParams struct {
	Addresses []string `json:"addresses"`
}

// RpcGetAddressTxIdsResp contains RPC response :: txid array or raw tx
type RpcGetAddressTxIdsResp struct {
	Result []string `json:"result"`
}

// RPCRawTxResponse contains RPC response :: raw transaction data
//type RPCRawTxResponse struct {
//	Result []string `json:"result"`
//}

// getData - ranges through transactions, returns txids or raw transactions
func getData(cmd string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		apiResp := api.Response{}

		var incomingTxs AddressesReq

		err := json.NewDecoder(r.Body).Decode(&incomingTxs)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnErr := api.AppRespErrors.ServerError
			returnErr.ErrorMessage = fmt.Sprintf("Server error: %v", err)
			apiResp.Errors = append(apiResp.Errors, returnErr)
			apiResp.Send(w)
			return
		}

		resp := buildResponse(incomingTxs)

		apiResp.Data = resp

		apiResp.Send(w)

		return
	})
}


func buildResponse( incomingAddreses AddressesReq ) Response {

	resp := Response{}



	for _, item := range incomingAddreses.Addresses {

		if item.Currency == "NAV" {

			result := Result{}
			result.Currency = "NAV"


			result.Addresses = getTransactionsForAddresses(item.Addresses)

			resp.Results = append(resp.Results, result)



		}

	}


	 return resp

}


func getTransactionsForAddresses(addresses []string) []Address {

	adds := []Address{}

	for _, addressStr := range addresses {

		addStruct := Address{}
		addStruct.Address = addressStr
		rpcTxIDsResp := getTxIDForAddressFromDaemon(addressStr)

		// for all the txIds from the rpc we need to create a transaction
		for _, txId := range rpcTxIDsResp.Result {

			trans := Transaction{Txid:txId}
			addStruct.Transactions = append(addStruct.Transactions, trans)

			//TODO: - get raw transaction from rpc
			//TODO: - get serialised transaction from rpc

		}

		adds = append(adds, addStruct)

	}

	return adds
}

// Gets all the transaction ids from the daemon for a given address
func getTxIDForAddressFromDaemon(address string) RpcGetAddressTxIdsResp {


	getParams := RPCGetAddressTxIDParams{}

	getParams.Addresses = append(getParams.Addresses, address)

	n := daemonrpc.RpcRequestData{}
	n.Method = "getaddresstxids"
	n.Params = []RPCGetAddressTxIDParams{getParams}

	resp, rpcErr := daemonrpc.RequestDaemon(n, conf.NavConf)

	if rpcErr != nil {
		//daemonrpc.RpcFailed(rpcErr, w, r)
	}

	rpcTxIdResults := RpcGetAddressTxIdsResp{}

	jsonErr := json.NewDecoder(resp.Body).Decode(&rpcTxIdResults)

	if jsonErr != nil {
		//returnErr := api.AppRespErrors.JSONDecodeError
		//returnErr.ErrorMessage = fmt.Sprintf("TxId JSON Decode Error: %v", jsonErr)
		//apiResp.Errors = append(apiResp.Errors, returnErr)
		//apiResp.Send(w)
	}



	return rpcTxIdResults

}




/*

// getDataFromAddresses returns txids from supplied addresses
func getDataFromAddresses(addresses []string, resp *Response) Response {



	//txout := OutgoingTransactions{}
	//txout.Currency = "NAV"


	for _, address := range addresses {
		txIDs := getDataFromAddress(address)
		txout.OutgoingAddresses = append(txout.OutgoingAddresses, createResponseObject(address, txIDs))
	}

	return txout

}

// getDataFromAddress issuces RPC calls, returns response (txid array)
func getDataFromAddress(address strings) RpcGetAddressTxIdsResp {

	apiResp := api.Response{}

	getParams := RPCGetAddressTxIDParams{}

	getParams.Addresses = append(getParams.Addresses, address)

	n := daemonrpc.RpcRequestData{}
	n.Method = "getaddresstxids"
	n.Params = []RPCGetAddressTxIDParams{getParams}

	resp, rpcErr := daemonrpc.RequestDaemon(n, conf.NavConf)

	if rpcErr != nil {
		//daemonrpc.RpcFailed(rpcErr, w, r)
	}

	txid := RpcGetAddressTxIdsResp{}

	jsonErr := json.NewDecoder(resp.Body).Decode(&txid)

	if jsonErr != nil {
		//returnErr := api.AppRespErrors.JSONDecodeError
		//returnErr.ErrorMessage = fmt.Sprintf("TxId JSON Decode Error: %v", jsonErr)
		//apiResp.Errors = append(apiResp.Errors, returnErr)
		//apiResp.Send(w)
	}

	//rawtx := RpcGetAddressTxIdsResp{}
	//
	//resp := txid
	//
	//if cmd == "raw" {
	//	for _, address := range txid.Result {
	//		rawtx.Result = getRawTransactionsFromTxId(txid.Result, w, r)
	//	}
	//	resp = append...
	//}

	return txid

}

// getRawTransactionFromTxId return the serialized, hex-encoded data for provided 'txid'
func getRawTransactionsFromTxId(txid string) RPCRawTxResponse {

	apiResp := api.Response{}

	n := daemonrpc.RpcRequestData{}
	n.Method = "getrawtransaction"
	n.Params = txid

	resp, rpcErr := daemonrpc.RequestDaemon(n, conf.NavConf)

	if rpcErr != nil {
		//daemonrpc.RpcFailed(rpcErr, w, r)
	}

	rawTx := RPCRawTxResponse{}

	jsonErr := json.NewDecoder(resp.Body).Decode(&rawTx)

	if jsonErr != nil {
		//returnErr := api.AppRespErrors.JSONDecodeError
		//returnErr.ErrorMessage = fmt.Sprintf("Raw Tx JSON Decode Error: %v", jsonErr)
		//apiResp.Errors = append(apiResp.Errors, returnErr)
		//apiResp.Send(w)
	}

	return rawTx

}



// createResponseObject formats the address, array of txids into outgoing address object
func createResponseObject(address string, txIDs RpcGetAddressTxIdsResp) OutgoingAddressObject {

	outAddObj := OutgoingAddressObject{}
	outAddObj.Address = address
	outAddObj.OutgoingTxIdsArray = txIDs.Result

	return outAddObj

}

*/
