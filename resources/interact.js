const ws = new WebSocket('ws://'+location.host+'/monitor')
const graph = document.getElementById('graph')
const btnPlay = document.getElementById('btn-play')
const btnPause = document.getElementById('btn-pause')
const annotations = []
const buttons = []
let signedon = false
let scrolling = true
let listened = false
let stopped = false
let timestamp = Date.now()
let interval = 100 // milliseconds

function relayout() {
   // if 'xaxis.range' is present and is a date, ignore automatic update
   if (!scrolling || typeof arguments[0]['xaxis.range'] !== 'undefined' && arguments[0]['xaxis.range'][0] instanceof Date)
      return;

   scrolling = false
   btnPlay.checked = false
   btnPlay.parentElement.className = 'btn btn-primary'
   btnPause.checked = true
   btnPause.parentElement.className = 'btn btn-primary active'
}

function refresh(msg) {
   let data = []

   for (const heading of msg.Enabled) {
      data.push({
         name: heading,
         type: msg.Enabled.length > 20 ? 'scattergl' : 'scatter',
         mode: 'lines',
         hoverlabel: {
            namelength: 100
         },
         x: [],
         y: []
      })
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
         orientation: msg.Enabled.length > 20 ? 'v' : 'h'
      }
   }

   Plotly.react(graph, data, layout, {displaylogo: false, responsive: true})

   // used to check if rangeslider should be updated or not
   if (!listened) {
      graph.on('plotly_relayout', relayout)
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
   if (scrolling)
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
   ws.send(val)
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
   const elem = document.getElementById('data-interval')
   elem.parentElement.nextSibling.data = ' '+data.Interval+'ms'
   elem.value = Math.log2(data.Interval)

   for (let i = 0; i < data.Tree.length; i++) {
      for (const key in data.Tree[i]) {
         if (!data.Tree[i].hasOwnProperty(key))
            continue

         subtree = document.createElement('details')
         let node = document.createElement('summary')
         subtree.appendChild(node)
         let text = document.createTextNode(key+' metrics')
         node.appendChild(text)

         elems = data.Tree[i][key]

         // special button to activate all events
         subtree.appendChild(button('all'))

         for (const elem of elems)
            subtree.appendChild(button(elem))

         let container = document.querySelector('#events')
         container.appendChild(subtree)
      }
   }
}

ws.onmessage = function(e) {
   let data = JSON.parse(e.data)

   if (signedon == false) {
      signon(data)
      signedon = true
      return
   }

   if (data.Op == 'data') {
      update(data)
   } else if (data.Op == 'enabled') {
      // handle JSON collapsing empty array
      if (data.Enabled == null)
         data.Enabled = []

      for (let btn of buttons)
         btn.className = data.Enabled.includes(btn.firstChild.nodeValue) ? 'btn btn-primary btn-sm m-1' : 'btn btn-light btn-sm m-1'

      refresh(data)
   } else if (data.Op == 'label')
      label(data)
   else if (data.Op == 'data')
      update(data)
   else
      console.log('unknown op '+data.Op)
}

ws.onopen = function(e) {
   ws.send('463ba1974b06')
}

ws.onerror = function(e) {
   console.log('error')
}

setInterval(scroll, interval)

function play() {
   if (stopped) {
      ws.send(JSON.stringify({Op: 'start'}))
      stopped = false
   }

   scrolling = true
}

function pause() {
   scrolling = false
}

function stop() {
   scrolling = false
   ws.send(JSON.stringify({Op: 'stop'}))
   stopped = true
}

function slider() {
   const val = Math.pow(2, Number(arguments[0].value))
   arguments[0].parentElement.nextSibling.data = ' '+val+'ms'
   const msg = JSON.stringify({Op: 'interval', Value: String(val)})
   ws.send(msg)
}
