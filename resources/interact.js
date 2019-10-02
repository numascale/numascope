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

'use strict';

const graph = document.getElementById('graph')
const btnPlay = document.getElementById('btn-play')
const btnPause = document.getElementById('btn-pause')
const btnStop = document.getElementById('btn-stop')
const radServerAverage = document.getElementById('serverAverage')
const radSocketAverage = document.getElementById('socketAverage')
const annotations = []
const buttons = []
let socketAverage = true
let socket
let signedon
let sources
let scrolling = true
let listened = false
let stopped = false
let discrete = false
let timestamp = Date.now()
let interval = 100 // milliseconds
let offline = false
let filter
const headings = []

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

   discrete = msg.Discrete
   radServerAverage.checked = !discrete

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
                  type: 'scatter',
                  mode: 'lines',
                  hoverlabel: {namelength: 50},
                  x: [], y: []
               })
            }
         } else {
            data.push({
               name: heading,
               type: 'scatter',
               mode: 'lines',
               hoverlabel: {namelength: 50},
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
      x: new Date(data.Timestamp / 1e3),
      y: 0,
      text: data.Label,
      arrowhead: 3,
      ax: 0,
      ay: 40
   })

   Plotly.relayout(graph, {annotations: annotations})
}

function update(data) {
   const indicies = []
   const x = []
   const y = []

   for (let i = 0; i < data[0].length-1; i++) {
      indicies.push(i)
      x.push([])
      y.push([])
   }

   // ensure graph scrolling is synchronised
   timestamp = data[data.length-1][0] / 1e3

   for (const update of data) {
      const time = new Date(update[0] / 1e3)

      for (let i = 1; i < update.length; i++) {
         x[i-1].push(time)
         y[i-1].push(update[i])
      }
   }

   Plotly.extendTraces(graph, {x: x, y: y}, indicies)
}

function scroll() {
   if (scrolling && listened)
      Plotly.relayout(graph, 'xaxis.range', [new Date(timestamp - 60e3), new Date(timestamp)])

   timestamp += interval
}

function select(info) {
   const msg = {
      Op: 'update',
      Event: info.target.innerText,
      State: info.target.className.includes('btn-primary') ? 'off' : 'on'
   }

   const val = JSON.stringify(msg)
   socket.send(val)
}

function button(name, on) {
   let btn = document.createElement('button')
   btn.onclick = select

   let text = document.createTextNode(name)
   btn.appendChild(text)
   btn.className = 'btn btn-light btn-sm m-1'

   if (on)
      btn.className += ' btn-primary'

   buttons.push(btn)

   return btn
}

function filterUNC(elems) {
   if (!radSocketAverage.checked)
      return elems

   // map input series to output series
   filter = []
   const headings = []

   for (let i = 0; i < elems.length; i++) {
      const name = elems[i].replace(/\.\d+/, '')

      // coalesce with any matching heading
      const j = headings.indexOf(name)

      if (j == -1) {
         filter.push(headings.length)
         headings.push(name)
      } else
         filter.push(j)
   }

   return headings
}

function filterNC2(elems) {
   let expr

   if (radServerAverage.checked && !radSocketAverage.checked)
      expr = /:\d+/
   else if (radServerAverage.checked && radServerAverage.checked)
      expr = /( \d+)?:\d+/
   else if (!radServerAverage.checked && !radServerAverage.checked)
      expr = / \d+/
   else // neither set
      return elems

   if (!radServerAverage.checked)
      return elems

   // map input series to output series
   filter = []
   const headings = []

   for (let i = 0; i < elems.length; i++) {
      const name = elems[i].replace(expr, '')

      // coalesce with any matching heading
      const j = headings.indexOf(name)

      if (j == -1) {
         filter.push(headings.length)
         headings.push(name)
      } else
         filter.push(j)
   }

   return headings
}

// takes an array, reduces it with filter[] and returns the result
function reduce(ents) {
   if (typeof filter === 'undefined')
      return ents

   const out = []

   for (let col = 0; col < ents.length; col++) {
      const dest = filter[col]

      if (typeof out[dest] === 'undefined')
         out[dest] = ents[col]
      else
         out[dest] += ents[col]
   }

   return out
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
      let elems = data.Tree[key]

      if (key == 'UNC')
         elems = filterUNC(elems)

      const subtree = document.createElement('details')
      const node = document.createElement('summary')
      subtree.appendChild(node)
      const text = document.createTextNode(key+' metrics')
      node.appendChild(text)

      // special button to activate all events
      subtree.appendChild(button('all', false))

      for (const elem of elems)
         subtree.appendChild(button(elem, false))

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

   if (data.Op == 'enabled')
      enabled(data)
   else if (data.Op == 'label')
      label(data)
   else
      update(data)
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
   if (typeof socket !== 'undefined')
      socket.send(JSON.stringify({Op: 'stop'}))
   stopped = true
}

function slider() {
   const val = Math.pow(2, Number(arguments[0].value))
   arguments[0].parentElement.nextSibling.data = ' '+val+'ms'
   const msg = JSON.stringify({Op: 'interval', Value: String(val)})
   socket.send(msg)
}

function serverAverageChange(control) {
   const val = control.checked
   const msg = JSON.stringify({Op: 'averaging', Value: String(val)})

   if (typeof socket !== 'undefined')
      socket.send(msg)
}

function socketAverageChange(control) {
   socketAverage = control.checked
}

function parse(file) {
   let json

   try {
      json = JSON.parse(file.target.result)
   } catch (e) {
      alert('Only valid JSON input supported')
      return
   }

   const data = []
   const total = json[0].length

/*   const grouping = document.getElementById('grouping')
   while (grouping.firstChild)
      grouping.removeChild(grouping.firstChild) */

   let headings = json[0].slice(1, total)

   switch(json[0][0]) {
   case "UNC":
      headings = filterUNC(headings)
//      grouping.appendChild(button('PE unit'))
      break
   case "Numascale NumaConnect2":
      headings = filterNC2(headings)
      break
   }

   const container = document.querySelector('#events')

   // remove any pre-existing sources from last session
   while (container.firstChild)
      container.removeChild(container.firstChild)

   const subtree = document.createElement('details')
   const node = document.createElement('summary')
   subtree.appendChild(node)
   const text = document.createTextNode(json[0][0]+' metrics')
   node.appendChild(text)

   // special button to activate all events
   subtree.appendChild(button('all', false))

   for (const heading of headings) {
      data.push({
         name: heading,
         type: 'scatter',
         mode: 'lines',
         hoverlabel: {namelength: 50},
         x: [], y: []
      })

      subtree.appendChild(button(heading, true))
   }

   container.appendChild(subtree)

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
      },
      annotations: []
   }

   for (let row = 1; row < json.length; row++) {
      const val = json[row][0]

      // handle general commands
      if (isNaN(val)) {
         switch(val) {
         case 'label':
            layout.annotations.push({
               x: new Date(json[row][1] / 1e3),
               y: 0,
               text: json[row][2],
               arrowhead: 3,
               ax: 0,
               ay: 40
            })
            break;
         default:
            alert('unknown op '+op)
         }

         continue
      }

      const time = new Date(val / 1e3)
      const elems = reduce(json[row].slice(1, json[row].length))

      for (let elem = 0; elem < elems.length; elem++) {
         data[elem].x.push(time)
         data[elem].y.push(elems[elem])
      }
   }

   Plotly.react(graph, data, layout, {displaylogo: false, responsive: true}})
}

function load(file) {
   btnPlay.checked = false
   btnPlay.parentElement.className = 'btn btn-primary'
   btnPause.checked = false
   btnPause.parentElement.className = 'btn btn-primary'
   btnStop.checked = true
   btnStop.parentElement.className = 'btn btn-primary active'

   stop()

   const reader = new FileReader()
   reader.onload = parse
   reader.readAsText(file)
   document.title = file.name+' - numascope'
}

if (location.host == '' || location.protocol == 'https:') {
   document.getElementById('btn-play').parentElement.className += ' disabled'
   document.getElementById('btn-pause').parentElement.className += ' disabled'
   document.getElementById('btn-stop').parentElement.className += ' disabled'
//   radServerAverage.disabled = true
   document.getElementById('data-interval').disabled = true
   document.getElementById('loading').innerHTML = 'Standalone mode'
   offline = true
} else
   connect()
