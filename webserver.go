package main

import (
   "fmt"
   "log"
   "net/http"
   "github.com/gorilla/websocket"
)

var (
   upgrader = websocket.Upgrader{}
   connections []*websocket.Conn
)

type SignonMessage struct {
   Interval int
   Tree     []map[string][]string
}

func monitor(w http.ResponseWriter, r *http.Request) {
   c, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
      log.Print("upgrade:", err)
      return
   }

   defer c.Close()

   // handshake
   _, message, err := c.ReadMessage()
   if err != nil {
      log.Println("read:", err)
      return
   }

   if string(message) != "463ba1974b06" {
      log.Println("auth failed")
      return
   }

   log.Println("auth succeeded")

   msg := SignonMessage{Interval: interval, Tree: make([]map[string][]string, len(sensors))}

   for i, sensor := range sensors {
      msg.Tree[i] = make(map[string][]string)
      name := sensor.sensor.name()
      supported := *sensor.sensor.supported()

      msg.Tree[i][name] = make([]string, len(supported))

      for j, val := range supported {
         msg.Tree[i][name][j] = val.desc
      }
   }

   err = c.WriteJSON(&msg)
   if err != nil {
      log.Println("write:", err)
      return
   }

   connections = append(connections, c)

   for {
   }
}

func initweb(addr string) {
   fileServer := http.FileServer(http.Dir("resources"))
   http.Handle("/", fileServer)
   http.HandleFunc("/monitor", monitor)

   go http.ListenAndServe(addr, nil)
   fmt.Printf("interface available at http://%v\n", addr)
}
