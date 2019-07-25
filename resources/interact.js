/*  Copyright (C) 2019 Daniel J Blueman
    This file is part of Numascope.

    Numascope is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    Numascope is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with Numascope.  If not, see <https://www.gnu.org/licenses/>.
*/

const graph = document.getElementById('graph')
const btnPlay = document.getElementById('btn-play')
const btnPause = document.getElementById('btn-pause')
const annotations = []
const buttons = []
let socket
let signedon
let sources
let scrolling = true
let listened = false
let stopped = false
let discrete = false
let timestamp = Date.now()
let interval = 100 // milliseconds

function connect() {
   socket = new WebSocket('ws://'+location.host+'/monitor')

   socket.onmessage = receive
   socket.onopen = function(e) {
      signedon = false
      socket.send('463ba1974b06')
   }

   socket.onclose = function(e) {
      $('#connecting').show()
   }
}

function subset(set, val) {
   for (const sensor in set) {
      if (set[sensor].includes(val))
         return true
   }

   return false
}

function relayout() {
   // if 'xaxis.range' is present and is a date, ignore automatic update
   if (!scrolling || typeof arguments[0]['xaxis.range'] !== 'undefined' && arguments[0]['xaxis.range'][0] instanceof Date || arguments[0]['autosize'] !== 'undefined')
      return;

   scrolling = false
   btnPlay.checked = false
   btnPlay.parentElement.className = 'btn btn-primary'
   btnPause.checked = true
   btnPause.parentElement.className = 'btn btn-primary active'
}

function enabled(msg) {
   var elem = document.getElementById('data-interval')
   elem.parentElement.nextSibling.data = ' '+msg.Interval+'ms'
   elem.value = Math.log2(msg.Interval)

   var elem = document.getElementById('averaging')
   discrete = msg.Discrete
   elem.checked = !discrete

   for (let btn of buttons)
      btn.className = subset(msg.Enabled, btn.firstChild.nodeValue) ? 'btn btn-primary btn-sm m-1' : 'btn btn-light btn-sm m-1'

   let data = []
   let total = 0

   for (const sensor in msg.Enabled)
      total += msg.Enabled[sensor].length * (discrete ? sources[sensor] : 1)


   for (const sensor in msg.Enabled) {
      for (const heading of msg.Enabled[sensor]) {
         if (discrete && sources[sensor] > 1) {
            for (let i = 0; i < sources[sensor]; i++) {
               data.push({
                  name: heading+':'+i,
                  type: total > 20  ? 'scattergl' : 'scatter',
                  mode: 'lines',
                  hoverlabel: {namelength: 100},
                  x: [], y: []
               })
            }
         } else {
            data.push({
               name: heading,
               type: total > 20  ? 'scattergl' : 'scatter',
               mode: 'lines',
               hoverlabel: {namelength: 100},
               x: [], y: []
            })
         }
      }
   }

   const layout = {
      autosize: true,
      height: 700,
      xaxis: {
         rangeslider: {}
      },
      yaxis: {
         title: 'events'
      },
      legend: {
         yanchor: 'top',
         y: -0.5,
         orientation: total > 20 ? 'v' : 'h'
      }
   }

   Plotly.react(graph, data, layout, {displaylogo: false, responsive: true})

   // used to check if rangeslider should be updated or not
   if (!listened) {
      graph.on('plotly_relayout', relayout)
      setInterval(scroll, interval)
      listened = true
   }
}

function label(data) {
   annotations.push({
      x: new Date(data.Timestamp),
      y: 0,
      text: data.Label,
      arrowhead: 3,
      ax: 0,
      ay: 40
   })

   Plotly.relayout(graph, {annotations: annotations})
}

function update(data) {
   // handle JSON collapsing empty array
   if (data.Values == null)
      data.Values = []

   timestamp = data.Timestamp

   let update = {
      x: [],
      y: []
   }

   let indicies = []
   const time = new Date(timestamp)

   for (let i = 0; i < data.Values.length; i++) {
      update.x.push([time])
      update.y.push([data.Values[i]])
      indicies.push(i)
   }

   Plotly.extendTraces(graph, update, indicies)
}

function scroll() {
   if (scrolling && listened)
      Plotly.relayout(graph, 'xaxis.range', [new Date(timestamp-60000), new Date(timestamp)])

   timestamp += interval
}

function select(info) {
   const msg = {
      Op: 'update',
      Event: info.target.innerText,
      State: info.target.className.includes('btn-primary') ? 'off' : 'on'
   }

   val = JSON.stringify(msg)
   socket.send(val)
}

function button(name) {
   let btn = document.createElement('button')
   btn.onclick = select

   let text = document.createTextNode(name)
   btn.appendChild(text)
   btn.className = 'btn btn-light btn-sm m-1'
   buttons.push(btn)

   return btn
}

function signon(data) {
   $('#connecting').hide()
   $('#loading').hide()

   sources = data.Sources
   const container = document.querySelector('#events')

   // remove any pre-existing sources from last session
   while (container.firstChild)
      container.removeChild(container.firstChild)

   for (const key in data.Tree) {
      subtree = document.createElement('details')
      let node = document.createElement('summary')
      subtree.appendChild(node)
      let text = document.createTextNode(key+' metrics')
      node.appendChild(text)
      elems = data.Tree[key]

      // special button to activate all events
      subtree.appendChild(button('all'))

      for (const elem of elems)
         subtree.appendChild(button(elem))

      container.appendChild(subtree)
   }
}

function receive(e) {
   let data = JSON.parse(e.data)

   if (signedon == false) {
      signon(data)
      signedon = true
      return
   }

   if (data.Op == 'data') {
      update(data)
   } else if (data.Op == 'enabled') {
      enabled(data)
   } else if (data.Op == 'label')
      label(data)
   else
      console.log('unknown op '+data)
}

function play() {
   if (stopped) {
      socket.send(JSON.stringify({Op: 'start'}))
      stopped = false
   }

   scrolling = true
}

function pause() {
   if (stopped) {
      socket.send(JSON.stringify({Op: 'start'}))
      stopped = false
   }

   scrolling = false
}

function stop() {
   scrolling = false
   socket.send(JSON.stringify({Op: 'stop'}))
   stopped = true
}

function slider() {
   const val = Math.pow(2, Number(arguments[0].value))
   arguments[0].parentElement.nextSibling.data = ' '+val+'ms'
   const msg = JSON.stringify({Op: 'interval', Value: String(val)})
   socket.send(msg)
}

function averaging() {
   const val = arguments[0].checked
   const msg = JSON.stringify({Op: 'averaging', Value: String(val)})
   socket.send(msg)
}

if (location.host == '') {
  document.getElementById('btn-play').parentElement.className += ' disabled'
  document.getElementById('btn-pause').parentElement.className += ' disabled'
  document.getElementById('btn-stop').parentElement.className += ' disabled'
  document.getElementById('averaging').disabled = true
  document.getElementById('data-interval').disabled = true
  document.getElementById('loading').innerHTML = 'Standalone mode'
} else
   connect()
