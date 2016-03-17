package main


import( "fmt"
        . "github.com/aerospike/aerospike-client-go"
)

func main() {
// remove timestamps from log messages
log.SetFlags(0)

// connect to the host
if conn, err := NewConnection("localhost:3000", 10*time.Second); err !=$
        log.Fatalln(err.Error())

} else {
        if infoMap, err := RequestInfo(conn, ""); err != nil {
                log.Fatalln(err.Error())
        } else {
                cnt := 1
                for k, v := range infoMap {
                        log.Printf("%d :  %s\n     %s\n\n", cnt, k, v)
                        cnt++
                }
        }
}
}
