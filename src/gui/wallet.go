// Wallet-related information for the GUI
package gui

import (
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/daemon"
	"github.com/skycoin/skycoin/src/wallet"
	"net/http"
	"strconv"
)

//var Wallets wallet.Wallets

/*
REFACTOR
*/

// Returns a *Balance

func GetWalletBalance(self *daemon.Gateway, walletID wallet.WalletID) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.GetWalletBalance(self.d.Visor.Visor, walletID)
	}
	r := <-self.Responses
	return r
}

// Returns map[WalletID]error

func SaveWallets(self *daemon.Gateway) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.SaveWallets(self.d.Visor.Visor)
	}
	r := <-self.Responses
	return r
}

// Returns error
func SaveWallet(self *daemon.Gateway, walletID wallet.WalletID) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.SaveWallet(self.d.Visor.Visor, walletID)
	}
	r := <-self.Responses
	return r
}

// Returns an error
func ReloadWallets(self *daemon.Gateway) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.ReloadWallets(self.d.Visor.Visor)
	}
	r := <-self.Responses
	return r
}

// Returns a *visor.ReadableWallet

func GetWallet(self *daemon.Gateway, walletID wallet.WalletID) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.GetWallet(self.d.Visor.Visor, walletID)
	}
	r := <-self.Responses
	return r
}

// Returns a *ReadableWallets

func GetWallets(self *daemon.Gateway) interface{} {
	self.Requests <- func() interface{} {
		return self.Visor.GetWallets(self.d.Visor.Visor)
	}
	r := <-self.Responses
	return r
}

// Returns a *ReadableWallet
// Deprecate

func CreateWallet(self *daemon.Gateway, seed string) interface{} {

	//w := v.CreateWallet()
	//return wallet.NewReadableWallet(w)

	//
	self.Requests <- func() interface{} {
		return self.Visor.CreateWallet(self.d.Visor.Visor)
	}
	r := <-self.Responses
	return r
	//
}

/*
REFACTOR
*/

// Returns the wallet's balance, both confirmed and predicted.  The predicted
// balance is the confirmed balance minus the pending spends.
func walletBalanceHandler(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		SendOr404(w, GetWalletBalance(gateway, wallet.WalletID(id)))
	}
}

// Creates and broadcasts a transaction sending money from one of our wallets
// to destination address.
func walletSpendHandler(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		walletId := wallet.WalletID(r.FormValue("id"))
		if walletId == "" {
			Error400(w, "Missing wallet_id")
			return
		}
		sdst := r.FormValue("dst")
		if sdst == "" {
			Error400(w, "Missing destination address \"dst\"")
			return
		}
		dst, err := cipher.DecodeBase58Address(sdst)
		if err != nil {
			Error400(w, "Invalid destination address")
			return
		}
		sfee := r.FormValue("fee")
		fee, err := strconv.ParseUint(sfee, 10, 64)
		if err != nil {
			Error400(w, "Invalid \"fee\" value")
			return
		}
		scoins := r.FormValue("coins")
		shours := r.FormValue("hours")
		coins, err := strconv.ParseUint(scoins, 10, 64)
		if err != nil {
			Error400(w, "Invalid \"coins\" value")
			return
		}
		hours, err := strconv.ParseUint(shours, 10, 64)
		if err != nil {
			Error400(w, "Invalid \"hours\" value")
			return
		}
		SendOr404(w, Spend(gateway, walletId, wallet.NewBalance(coins, hours),
			fee, dst))
	}
}

// Create a wallet if no ID provided.  Otherwise update an existing wallet.
// Name the wallet with "name".
func walletCreate(gateway *daemon.Gateway) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("API request made to create a wallet")
		//id := wallet.WalletID(r.FormValue("id"))
		name := r.FormValue("name")
		seed := r.FormValue("seed")

		// Create wallet
		//iw := gateway.CreateWallet("") //returns wallet
		//iw := wallet.NewReadableWallet(w)

		iw := wallet.NewReadableWallet(w)

		if iw != nil {
			w := iw.(wallet.Wallet)
			w.SetName(name)
			if err := SaveWallet(gateway, w.GetID()); err != nil {
				m := "Failed to save wallet after renaming: %v"
				logger.Critical(m, err)
			}
		}
		SendOr500(w, iw)
	}
}

func walletUpdate(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Update wallet
		id := wallet.WalletID(r.FormValue("id"))
		name := r.FormValue("name")
		iw := GetWallet(gateway, id)
		if iw != nil {
			w := iw.(wallet.Wallet)
			w.SetName(name)
			if err := SaveWallet(gateway, w.GetID()); err != nil {
				m := "Failed to save wallet after renaming: %v"
				logger.Critical(m, err)
			}
		}
		SendOr404(w, iw)
	}
}

// Returns a wallet by ID if GET.  Creates or updates a wallet if POST.
func walletGet(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			ret := GetWallet(gateway, wallet.WalletID(r.FormValue("id")))
			SendOr404(w, ret)
		}
	}
}

// Returns all loaded wallets
func walletsHandler(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//ret := wallet.Wallets.ToPublicReadable()
		ret := GetWallets(gateway)
		SendOr404(w, ret)
	}
}

// Saves all loaded wallets
func walletsSaveHandler(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errs := SaveWallets(gateway).(map[wallet.WalletID]error)
		if len(errs) != 0 {
			err := ""
			for id, e := range errs {
				err += string(id) + ": " + e.Error()
			}
			Error500(w, err)
		}
	}
}

// Loads/unloads wallets from the wallet directory
func walletsReloadHandler(gateway *daemon.Gateway) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := ReloadWallets(gateway)
		if err != nil {
			Error500(w, err.(error).Error())
		}
	}
}

func RegisterWalletHandlers(mux *http.ServeMux, gateway *daemon.Gateway) {
	// Returns wallet info
	// GET Arguments:
	//      id - Wallet ID.

	//   Creates a new wallet if no id given.  Will be assigned name if present.
	mux.HandleFunc("/wallet", walletGet(gateway))

	// POST/GET Arguments:
	//      name [optional]
	//		seed [optional]
	//create new wallet
	mux.HandleFunc("/wallet/create", walletCreate(gateway))

	//update an existing wallet
	mux.HandleFunc("/wallet/update", walletUpdate(gateway))

	// Returns the confirmed and predicted balance for a specific wallet.
	// The predicted balance is the confirmed balance minus any pending
	// spent amount.
	// GET arguments:
	//      id: Wallet ID
	mux.HandleFunc("/wallet/balance", walletBalanceHandler(gateway))

	// Sends coins&hours to another address.
	// POST arguments:
	//  id: Wallet ID
	//  coins: Number of coins to spend
	//  hours: Number of hours to spends
	//  fee: Number of hours to use as fee, on top of the default fee.
	//  Returns total amount spent if successful, otherwise error describing
	//  failure status.
	mux.HandleFunc("/wallet/spend", walletSpendHandler(gateway))

	// Returns all loaded wallets
	mux.HandleFunc("/wallets", walletsHandler(gateway))
	// Saves all wallets to disk. Returns nothing if it works. Otherwise returns
	// 500 status with error message.

	mux.HandleFunc("/wallets/save", walletsSaveHandler(gateway))
	// Rescans the wallet directory and loads/unloads wallets based on which
	// files are present. Returns nothing if it works. Otherwise returns
	// 500 status with error message.
	mux.HandleFunc("/wallets/reload", walletsReloadHandler(gateway))
}
