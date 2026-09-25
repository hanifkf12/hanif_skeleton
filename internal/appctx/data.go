package appctx

// Data membawa request HTTP yang masuk ke usecase.
//
// Request adalah satu-satunya cara usecase mengakses request, sehingga
// business layer tidak terikat pada framework HTTP mana pun.
type Data struct {
	Request Request
}
