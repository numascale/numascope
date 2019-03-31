package main

import (
   //   "html/template"
   "encoding/json"
   "fmt"
   "github.com/gorilla/websocket"
   "log"
   "math/rand"
   "net/http"
   "time"
)

var (
   upgrader = websocket.Upgrader{} // use default options
)

type Message struct {
   DeviceID    string
   State       string
   Temperature int
}

func monitor(w http.ResponseWriter, r *http.Request) {
   c, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
      log.Print("upgrade:", err)
      return
   }

   defer c.Close()

   // handshake
   msgType, message, err := c.ReadMessage()
   if err != nil {
      log.Println("read:", err)
      return
   }
   if string(message) != "463ba1974b06" {
      log.Println("auth failed")
      return
   }

   log.Println("auth succeeded")

   for {
      t := int(rand.Float32() * 70)
      message := Message{"device1234", "sleeping", t}

      bytes, err := json.Marshal(message)
      if err != nil {
         log.Println("marshall:", err)
         break
      }

      err = c.WriteMessage(msgType, bytes)
      if err != nil {
         log.Println("write:", err)
         break
      }

      // log.Println("wrote")
      time.Sleep(1000 * time.Millisecond)
   }
}

func initweb(addr string) {
   fileServer := http.FileServer(http.Dir("resources"))
   http.Handle("/", fileServer)
   http.HandleFunc("/monitor", monitor)

   go http.ListenAndServe(addr, nil)
   fmt.Printf("Interface available at http://%v\n", addr)
}
