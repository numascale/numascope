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

const defaultTraces = {
   NumaConnect2: '% wait cycles',
   UNC: 'IOA SCI Intr'
}

const graph = document.getElementById('graph')
const btnPlay = document.getElementById('btn-play')
const btnPause = document.getElementById('btn-pause')
const btnStop = document.getElementById('btn-stop')
const radServerGroup = document.getElementById('serverGroup')
const radUnitGroup = document.getElementById('unitGroup')
const annotations = []
const buttons = []
let normalise // used to derive percentage
let unitGroup = true
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
const layout = {
   height: 500,
   xaxis: {
      title: 'seconds',
      rangeslider: {},
      hoverformat: ',.3s'
   },
   yaxis: {
      title: 'events',
      hoverformat: ',.3s',
      titlefont: {color: '#1f77b4'},
      tickfont: {color: '#1f77b4'},
      rangemode: 'tozero'
   },
   yaxis2: {
      title: '%',
      hoverformat: ',.3r',
      titlefont: {color: '#ff7f0e'},
      tickfont: {color: '#ff7f0e'},
      overlaying: 'y',
      side: 'right',
      showgrid: false,
      rangemode: 'tozero'
   },
   legend: {
      borderwidth: 34,
      bordercolor: "#ffffff"
   },
   annotations: []
}

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
   radServerGroup.checked = !discrete

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
                  hoverlabel: {namelength: 80},
                  x: [], y: [],
                  yaxis: heading[0] == '%' ? 'y2' : 'y1',
                  visible: heading.includes(defaultTraces[technology]) ? 'true' : 'legendonly'
               })
            }
         } else {
            data.push({
               name: heading,
               type: 'scatter',
               mode: 'lines',
               hoverlabel: {namelength: 80},
               x: [], y: [],
               yaxis: (heading[0] == '%') ? 'y2' : 'y1',
               visible: heading.includes(defaultTraces[technology]) ? 'true' : 'legendonly'
            })
         }
      }
   }

   layout.legend.orientation = total > 20 ? 'v' : 'h'
   Plotly.react(graph, data, layout, {displaylogo: false, responsive: true})

   // used to check if rangeslider should be updated or not
   if (!listened) {
      graph.on('plotly_relayout', relayout)
      setInterval(scroll, interval)
      listened = true
   }
}

function label(elem) {
   annotations.push({
      x: new Date(elem.Timestamp / 1e3),
      y: 0,
      text: elem.Label,
      arrowhead: 3,
      ax: 0,
      ay: 40
   })

   Plotly.relayout(graph, {annotations: annotations})
}

function update(elem) {
   const indicies = []
   const x = []
   const y = []

   for (let i = 0; i < elem[0].length-1; i++) {
      indicies.push(i)
      x.push([])
      y.push([])
   }

   // ensure graph scrolling is synchronised
   timestamp = elem[elem.length-1][0] / 1e3

   for (const update of elem) {
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

function filterGen(elems, exprs) {
   if (!exprs.length)
      return elems

   // map input series to output series
   filter = []
   const headings = []

   for (let i = 0; i < elems.length; i++) {
      let name = elems[i]

      for (const expr of exprs)
         name = name.replace(expr, '')

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

function filterUNC(elems) {
   let exprs = []

   if (radServerGroup.checked)
      exprs.push(/UNC\d+ /)

   if (radServerGroup.checked)
      exprs.push(/\.\d+/)

   return filterGen(elems, exprs)
}

function filterNC2(elems) {
   let exprs = []

   if (radServerGroup.checked)
      exprs.push(/:\d+/)

   if (radUnitGroup.checked)
      exprs.push(/ \d+/)

   return filterGen(elems, exprs)
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

function signon(elem) {
   $('#connecting').hide()
   $('#loading').hide()

   sources = elem.Sources
   reset()

   const container = document.querySelector('#events')

   for (const key in elem.Tree) {
      let elems = elem.Tree[key]

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
   let input = JSON.parse(e.data)

   if (signedon == false) {
      signon(input)
      signedon = true
      return
   }

   if (input.Op == 'enabled')
      enabled(input)
   else if (input.Op == 'label')
      label(input)
   else
      update(input)
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

function serverGroupChange(control) {
   const val = control.checked
   const msg = JSON.stringify({Op: 'averaging', Value: String(val)})

   if (typeof socket !== 'undefined')
      socket.send(msg)
}

function unitGroupChange(control) {
   unitGroup = control.checked
}

// remove any pre-existing sources from last session
function reset() {
   for (const container of [document.querySelector('#events'), document.getElementById('totals')])
      while (container.firstChild)
         container.removeChild(container.firstChild)
}

function parse(file) {
   let json

   try {
      json = JSON.parse(file.target.result)
   } catch (e) {
      alert('Input file is not well-formed JSON\n\n'+e)
      return
   }

   const data = []
   const total = json[0].length

/*   const grouping = document.getElementById('grouping')
   while (grouping.firstChild)
      grouping.removeChild(grouping.firstChild) */

   const technology = json[0][0]
   let headings = json[1]

   switch(technology) {
   case 'UNC':
      headings = filterUNC(headings)
//      grouping.appendChild(button('PE unit'))
      break
   case 'NumaConnect2':
      headings = filterNC2(headings)
      break
   }

   normalise = json[0][2] / 100
   reset()

   const subtree = document.createElement('details')
   const node = document.createElement('summary')
   subtree.appendChild(node)
   const text = document.createTextNode(technology+' metrics')
   node.appendChild(text)

   // special button to activate all events
   subtree.appendChild(button('all', false))

   const totals = []

   for (const heading of headings) {
      data.push({
         name: heading,
         type: 'scatter',
         mode: 'lines',
         hoverlabel: {namelength: 80},
         x: [], y: [],
         yaxis: (heading[0] == '%') ? 'y2' : 'y1',
         visible: heading.includes(defaultTraces[technology]) ? 'true' : 'legendonly'
      })

      subtree.appendChild(button(heading, true))
      totals.push(0)
   }

   const container = document.querySelector('#events')
   container.appendChild(subtree)

   const timeOffset = json[2][0]
   for (let row = 2; row < json.length; row++) {
      const val = json[row][0]

      // handle general commands
      if (isNaN(val)) {
         switch(val) {
         case 'label':
            layout.annotations.push({
               x: (json[row][1] - timeOffset) / 1e6,
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

      const seconds = (val - timeOffset) / 1e6
      const elems = reduce(json[row].slice(1, json[row].length))

      for (let elem = 0; elem < elems.length; elem++) {
         data[dataOffset+elem].x.push(seconds)
         data[dataOffset+elem].y.push(
            (headings[elem][0] == '%') ? (elems[elem] / normalise) : elems[elem])
         totals[elem] += elems[elem]
      }
   }

   const totalsTable = document.getElementById('totals')
   const interval = (json[json.length-1][0] - json[2][0]) / 1e6
   document.getElementById('tableCaption').innerHTML = 'Total time '+interval.toFixed(2)+'s'
   let i = 0

   for (const heading of headings) {
      const row = totalsTable.insertRow(-1)
      const cell = row.insertCell(-1)

      cell.innerHTML = heading
      row.insertCell(-1).innerHTML = totals[i]
      row.insertCell(-1).innerHTML = Math.round(totals[i]/interval)

      i++
   }

   layout.legend.orientation = headings.length > 20 ? 'v' : 'h'
   Plotly.react(graph, data, layout, {displaylogo: false, responsive: true})
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
//   radServerGroup.disabled = true
   document.getElementById('data-interval').disabled = true
   document.getElementById('loading').innerHTML = 'Standalone mode'
   offline = true
} else
   connect()
