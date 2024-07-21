package constcoe

const (
	DIFFICUTY           = 1    //该值越大找到一个符合要求的值的难度越大，工作量越大
	INITCOIN            = 1000 //初始时总的bitcoin数目
	TRANSACTIONPOOLFILE = "./tmp/transaction_pool.data"
	BCPATH              = "./tmp/blocks"
	BCFILE              = "./tmp/blocks/MANIFEST"
	PREVHASH            = "这是创世区块"
	CHECKSUMLEN         = 4
	NETWORKVERSION      = byte(0x00)
	WALLETSDIR          = "./tmp/wallets/"
	WALLETSREFLIST      = "./tmp/ref_list/"
)
