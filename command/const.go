package command

var (
	dbSource       = 0
	dbDest         = 0
	con1           = 3 // concurrency for comparing
	con2           = 4 // concurrency for restoring
	sourceUsername = ""
	sourcePassword = ""
	hostDest       = "localhost:6379"
	destUsername   = ""
	destPassword   = ""
	hostSource     = "localhost:6380"
)
