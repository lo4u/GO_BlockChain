package cli

import (
	"GOPreject/blockchain"
	"GOPreject/utils"
	"GOPreject/wallet"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
}

func (pCLI *CommandLine) printUsage() {
	fmt.Println("Welcome to u1h's tiny blockchain system, usage is as follows:")
	fmt.Println("------------------------------------------------------------------------------------------------------------------")
	fmt.Println("All you need is to first create a wallet.")
	fmt.Println("And then you can use the wallet's address to create a blockchain and declare the owner.")
	fmt.Println("Make transactions to expand the blockchain.")
	fmt.Println("In addition, don't forget to run mine function after transactions are collected.")
	fmt.Println("Please make sure the UTXO set init before querying the balance of a wallet")
	fmt.Println("------------------------------------------------------------------------------------------------------------------")
	fmt.Println("createwallet -refname REFNAME				---->Creates and save a wallet. The refname is optional.")
	fmt.Println("walletinfo -refname NAME -address ADDRESS				---->Print the information of a wallet. At least one of the refname and address is supplied.")
	fmt.Println("walletupdate					---->Register and update all the wallets (especially when you add an existed .wlt file).")
	fmt.Println("walletlist					---->List all the wallets found (make sure you have run walletupdate first).")
	fmt.Println("createblockchain -refname NAME -address ADDRESS					---->Creates a blockchain with the owner you input (address or refname).")
	fmt.Println("initutxoset					---->Init all the UTXO sets of known wallets")
	fmt.Println("balance -refname REFNAME -address ADDRESS					---->Query the balance of the address or refname you input")
	fmt.Println("blockchaininfo							---->Prints the blocksin the blockchain")
	fmt.Println("sendbyname -from FROMNAME -to TONAME -amount AMOUNT		---->Make a transaction and put it into candidate block, by refname")
	fmt.Println("send -from FROMADDRESS -to TOADDRESS -amount AMOUNT					----Make a transaction and put it into candidate block")
	fmt.Println("mine								---->Mine and add a block to the chain")
	fmt.Println("------------------------------------------------------------------------------------------------------------------")
}

func (pCLI *CommandLine) createWallet(refName string) {
	//可能已有用户钱包列表了
	refList := wallet.LoadRefList()
	w := wallet.NewWallet()
	refList.BindRef(string(w.Address()), refName)
	w.Save()
	refList.Save()
	fmt.Println("Succeed in creating a wallet")
}

func (pCLI *CommandLine) walletInfo_refName(refName string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindAddress(refName)
	utils.Handle(err)
	pCLI.walletInfo_address(address)
}

func (*CommandLine) walletInfo_address(address string) {
	refList := wallet.LoadRefList()
	w := wallet.LoadWallet([]byte(address))

	fmt.Printf("Wallet address: %s\n", address)
	fmt.Printf("Wallet Public Key: %x\n", w.PublicKey)
	fmt.Printf("Wallet RefName: %s\n", (*refList)[address])
}

func (*CommandLine) walletUpdate() {
	refList := wallet.LoadRefList()
	refList.Update()
	refList.Save()
	fmt.Println("Succeed in updating wallets")
}

func (*CommandLine) walletList() {
	refList := wallet.LoadRefList()
	for address := range *refList {
		w := wallet.LoadWallet([]byte(address))
		fmt.Println("------------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Wallet address: %s\n", address)
		fmt.Printf("Wallet Public Key: %x\n", w.PublicKey)
		fmt.Printf("Wallet RefName: %s\n", (*refList)[address])
		fmt.Println("------------------------------------------------------------------------------------------------------------------")
		fmt.Println()
	}
}

func (*CommandLine) initUTXOSet() {
	pBlockChain := blockchain.ContinueBlockChain()
	defer pBlockChain.Database.Close()
	pRefList := wallet.LoadRefList()
	for addr := range *pRefList {
		utxoset := pBlockChain.CreataUTXOSet([]byte(addr))
		utxoset.DB.Close()
	}
	fmt.Println("Succeed in initializing UTXO sets")
}

func (pCLI *CommandLine) createBlockChain(address string) {
	pChain := blockchain.InitBlockChain([]byte(address))
	defer pChain.Database.Close()
	fmt.Println("Finish creatingn blockchain, and the owner is", address)
}

func (pCLI *CommandLine) createBlockChain_refName(refName string) {
	//先获取地址
	refList := wallet.LoadRefList()
	address, err := refList.FindAddress(refName)
	utils.Handle(err)

	pCLI.createBlockChain(address)
}

//弃用
// func (pCLI *CommandLine) balance(address string) {
// 	pChain := blockchain.ContinueBlockChain()
// 	defer pChain.Database.Close()
// 	balance, _ := pChain.FindUTXOs([]byte(address))
// 	fmt.Printf("Address: %s, Balance: %d \n", address, balance)
// }

func (*CommandLine) balance(address string) {
	pWallet := wallet.LoadWallet([]byte(address))
	amount := pWallet.GetBalance()
	fmt.Printf("Address: %s, Balance: %d\n", address, amount)
}

func (pCLI *CommandLine) balance_refName(refName string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindAddress(refName)
	utils.Handle(err)

	pCLI.balance(address)
}

func (pCLI *CommandLine) getBlockChainInfo() {
	pChain := blockchain.ContinueBlockChain()
	defer pChain.Database.Close()

	iter := pChain.NewIterator()
	end := pChain.End()
	for !iter.Equal(end) {
		pBlock := iter.Next()
		fmt.Println("----------------------------------------------------")
		fmt.Printf("Timestamp: %d\n", pBlock.Timestamp)
		fmt.Printf("Previous hash: %x\n", pBlock.PrevHash)
		fmt.Printf("Height: %d\n", pBlock.Height)
		fmt.Printf("Transaction: %v\n", pBlock.Transactions)
		fmt.Printf("hash: %x\n", pBlock.Hash)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(pBlock.ValidatePoW()))
		fmt.Printf("MTree's root hash: %x\n", pBlock.MTree.RootNode.HashData)
		fmt.Println("----------------------------------------------------")
		fmt.Println()
	}
}

func (*CommandLine) send(from, to string, amount int) {
	pChain := blockchain.ContinueBlockChain()
	defer pChain.Database.Close()

	pTx, isOK := pChain.CreateTransaction([]byte(from), []byte(to), amount)
	if !isOK {
		fmt.Println("Failed to create transaction.")
		return
	}
	pTP := blockchain.GetTransactionPool()
	pTP.AddTransaction(pTx)
	pTP.SaveFile()
	fmt.Println("Success!")
}

func (pCLI *CommandLine) send_refName(fromName, toName string, amount int) {
	//先获取地址
	refList := wallet.LoadRefList()
	from, err := refList.FindAddress(fromName)
	utils.Handle(err)
	to, err := refList.FindAddress(toName)
	utils.Handle(err)

	pCLI.send(from, to, amount)
}

func (pCLI *CommandLine) mine() {
	pChain := blockchain.ContinueBlockChain()
	defer pChain.Database.Close()

	ok := pChain.RunMine()
	fmt.Println("Finish Mining")
	if ok {
		pRefList := wallet.LoadRefList()
		for addr := range *pRefList {
			pChain.UpdateUTXOSet([]byte(addr))
		}
	}
	fmt.Println("Finish updating UTXO sets")
}

func (pCLI *CommandLine) validataArgs() {
	if len(os.Args) < 2 {
		pCLI.printUsage()
		runtime.Goexit()
	}
}
func (pCLI *CommandLine) Run() {
	pCLI.validataArgs()

	switch os.Args[1] {
	case "createwallet":
		flagSet := flag.NewFlagSet("createwallet", flag.ExitOnError)
		pRefName := flagSet.String("refname", "", "The refname of the wallet, optional")
		flagSet.Parse(os.Args[2:])
		pCLI.createWallet(*pRefName)
	case "walletinfo":
		flagSet := flag.NewFlagSet("walletinfo: at lease give one flag; if you give both two, i will pick refname", flag.ExitOnError)
		pAddress := flagSet.String("address", "", "The address corresponding to the wallet")
		pRefName := flagSet.String("refname", "", "The refname corresponding to the wallet")
		flagSet.Parse(os.Args[2:])
		if *pAddress == "" && *pRefName == "" {
			flagSet.Usage()
			runtime.Goexit()
		} else if *pRefName != "" {
			pCLI.walletInfo_refName(*pRefName)
		} else {
			pCLI.walletInfo_address(*pAddress)
		}
	case "walletupdate":
		pCLI.walletUpdate()
	case "walletlist":
		pCLI.walletList()
	case "createblockchain":
		flagSet := flag.NewFlagSet("createblockchain", flag.ExitOnError)
		pAddress := flagSet.String("address", "", "The address refer to the owner of the blockchain")
		pRefName := flagSet.String("refname", "", "The refname refer to the owner of the blockchain")
		flagSet.Parse(os.Args[2:])
		if *pAddress == "" && *pRefName == "" {
			flagSet.Usage()
			runtime.Goexit()
		} else if *pRefName != "" {
			pCLI.createBlockChain_refName(*pRefName)
		} else {
			pCLI.createBlockChain(*pAddress)
		}
	case "initutxoset":
		pCLI.initUTXOSet()
	case "balance":
		flagSet := flag.NewFlagSet("createblockchain", flag.ExitOnError)
		pAddress := flagSet.String("address", "", "Which address's balance you would query")
		pRefName := flagSet.String("refname", "", "Which refname's balance you would query")
		flagSet.Parse(os.Args[2:])
		if *pAddress == "" && *pRefName == "" {
			flagSet.Usage()
			runtime.Goexit()
		} else if *pRefName != "" {
			pCLI.balance_refName(*pRefName)
		} else {
			pCLI.balance(*pAddress)
		}
	case "blockchaininfo":
		pCLI.getBlockChainInfo()
	case "send":
		flagSet := flag.NewFlagSet("send", flag.ExitOnError)
		pFromAddress := flagSet.String("from", "", "whois address")
		pToAddress := flagSet.String("to", "", "whois address")
		pAmount := flagSet.Int("amount", 0, "whois address")
		flagSet.Parse(os.Args[2:])
		if *pFromAddress == "" || *pToAddress == "" {
			flagSet.Usage()
			runtime.Goexit()
		} else {
			pCLI.send(*pFromAddress, *pToAddress, *pAmount)
		}
	case "sendbyname":
		flagSet := flag.NewFlagSet("send", flag.ExitOnError)
		pFromName := flagSet.String("from", "", "Which refname does this record come from")
		pToName := flagSet.String("to", "", "Which refname does this record aim to")
		pAmount := flagSet.Int("amount", 0, "How much??")
		flagSet.Parse(os.Args[2:])
		if *pFromName == "" || *pToName == "" {
			flagSet.Usage()
			runtime.Goexit()
		} else {
			pCLI.send_refName(*pFromName, *pToName, *pAmount)
		}
	case "mine":
		pCLI.mine()
	default:
		pCLI.printUsage()
	}
}
