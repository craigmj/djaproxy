package django

import (
	"log"
	"os"
)

var elog = log.New(os.Stderr,  `E `, log.Lshortfile | log.LstdFlags)
var ilog = log.New(os.Stderr,  `I `, log.Lshortfile | log.LstdFlags)