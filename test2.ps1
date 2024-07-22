./init
./main.exe createwallet 
./main.exe walletlist
./main.exe createwallet -refname LeoCao
./main.exe walletinfo -refname LeoCao
./main.exe createwallet -refname Krad
./main.exe createwallet -refname Exia
./main.exe createwallet 
./main.exe walletlist
./main.exe createblockchain -refname LeoCao
./main.exe blockchaininfo
./main.exe initutxoset
./main.exe balance -refname LeoCao
./main.exe sendbyname -from LeoCao -to Krad -amount 100
./main.exe balance -refname Krad
./main.exe mine
./main.exe blockchaininfo
./main.exe balance -refname LeoCao
./main.exe balance -refname Krad
./main.exe sendbyname -from LeoCao -to Exia -amount 100
./main.exe sendbyname -from Krad -to Exia -amount 30
./main.exe mine
./main.exe blockchaininfo
./main.exe balance -refname LeoCao
./main.exe balance -refname Krad
./main.exe balance -refname Exia
./main.exe sendbyname -from Exia -to LeoCao -amount 90
./main.exe sendbyname -from Exia -to Krad -amount 90
./main.exe mine
./main.exe blockchaininfo
./main.exe balance -refname LeoCao
./main.exe balance -refname Krad
./main.exe balance -refname Exia