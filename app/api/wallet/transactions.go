package wallet

import (
	"io/ioutil"
	"net/http"

	"github.com/Encrypt-S/navpi-go/app/api"
	"github.com/Encrypt-S/navpi-go/app/conf"
	"github.com/Encrypt-S/navpi-go/app/daemon/daemonrpc"
	"github.com/gorilla/mux"
	"encoding/json"
	"fmt"
)

// InitTransactionHandlers sets up handlers for transaction-related rpc commands
func InitTransactionHandlers(r *mux.Router, prefix string) {

	namespace := "transactions"

	getAddressTxIdsPath := api.RouteBuilder(prefix, namespace, "v1", "getaddresstxids")
	api.OpenRouteHandler(getAddressTxIdsPath, r, getAddressTxIds())

}

// format of transactions json payload from incoming POST
//{"transactions": [
//{"currency":  "NAV", "address": "Nkjhsdfkjh834jdu"},
//{"currency":  "NAV", "address": "Nkjhsdfkjh834jdu"},
//{"currency":  "BTC", "address": "1kjhsdfkjh834jdu"},
//{"currency":  "BTC", "address": "Nkjhsdfkjh834jdu"}
//]}

// TODO: Decode top level transactions JSON array into a slice of structs

// first decode transactions json into a GetAddressTxIdsArray : Transactions slice
type GetAddressTxIdsArray struct {
	Transactions []GetAddressTxIdsJson `json:"array"`
}

// then iterate over the Transactions slice to get each GetAddressTxIdsJson
type GetAddressTxIdsJson struct {
	Currency   string  `json:"currency"`
	Address 	 string  `json:"address"`
}

// getAddressTxIds - executes "getaddresstxids" JSON-RPC command
// arguments - addresses array, start block height, end block height
// returns the txids for an address(es) (requires addressindex to be enabled).
func getAddressTxIds() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var getAddressTxIds GetAddressTxIdsArray
		apiResp := api.Response{}

		err := json.NewDecoder(r.Body).Decode(&getAddressTxIds)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			returnErr := api.AppRespErrors.ServerError
			returnErr.ErrorMessage = fmt.Sprintf("Server error: %v", err)
			apiResp.Errors = append(apiResp.Errors, returnErr)
			apiResp.Send(w)
			return
		}

		n := daemonrpc.RpcRequestData{}
		n.Method = "getaddresstxids"
		n.Params = getAddressTxIds.Transactions;

		resp, err := daemonrpc.RequestDaemon(n, conf.NavConf)

		if err != nil { // Handle errors requesting the daemon
			daemonrpc.RpcFailed(err, w, r)
			return
		}

		bodyText, err := ioutil.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(bodyText)

	})
}
