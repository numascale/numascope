const ws = new WebSocket('ws://'+location.host+'/monitor');
let signedon = false;
let buttons;
let first;

function graph(msg) {
   first = new Date(msg.Timestamp);
   let data = [];

   for (const heading of msg.Enabled) {
      data.push({
         name: heading,
         mode: 'lines',
         x: [],
         y: []
      });
   }

   const layout = {
      height: 700,
      xaxis: {
         rangeslider: {},
      },
      yaxis: {
         title: 'events',
      },
      layout: {
         "orientation": "h",
      }
   }

   Plotly.react('graph', data, layout);
}

function update(data) {
   // handle JSON collapsing empty array
   if (data.Values == null)
      data.Values = []

   const time = new Date(data.Timestamp);

   let update = {
      x: [],
      y: []
   }

   let indicies = [];

   for (let i = 0; i < data.Values.length; i++) {
      update.x.push([time]);
      update.y.push([data.Values[i]]);
      indicies.push(i);
   }

   const olderTime = time.setMinutes(time.getMinutes() - 1);
   const newerTime = time.setMinutes(time.getMinutes() + 1);
   const view = {
      xaxis: {
         type: 'date',
         range: [olderTime, newerTime],
         rangeslider: {}
     }
   };

   Plotly.extendTraces('graph', update, indicies)
   Plotly.relayout('graph', view);
}

function select(info) {
   const msg = {
      Op: "update",
      Event: info.target.innerText,
      State: info.target.className.includes('btn-primary') ? "off" : "on"
   };

   val = JSON.stringify(msg);
   ws.send(val);
}

function signon(data) {
   for (let i = 0; i < data.Tree.length; i++) {
      for (const key in data.Tree[i]) {
         if (!data.Tree[i].hasOwnProperty(key))
            continue;

         buttons = document.createElement('details');
         let node = document.createElement('summary');
         buttons.appendChild(node);
         let text = document.createTextNode(key+' metrics');
         node.appendChild(text);

         elems = data.Tree[i][key];

         for (const elem of elems) {
            let btn = document.createElement('button');
            btn.onclick = select;

            let text = document.createTextNode(elem);
            btn.appendChild(text);
            btn.className = 'btn btn-light btn-sm m-1';
            buttons.appendChild(btn);
         }

         let container = document.querySelector('#events');
         container.appendChild(buttons);
      }
   }
}

ws.onmessage = function(e) {
   let data = JSON.parse(e.data);

   if (signedon == false) {
      signon(data);
      signedon = true;
      return;
   }

   if (data.Op == 'enabled') {
      // handle JSON collapsing empty array
      if (data.Enabled == null)
         data.Enabled = []

      for (let btn of buttons.childNodes) {
         if (!btn.className.startsWith('btn')) {
            continue;
         }

         btn.className = data.Enabled.includes(btn.firstChild.nodeValue) ? 'btn btn-primary btn-sm m-1' : 'btn btn-light btn-sm m-1';
      }

      graph(data);
      return;
   }

   update(data);
}

ws.onopen = function(e) {
   ws.send('463ba1974b06')
}

ws.onclose = function(e) {
   console.log('closed');
}

ws.onerror = function(e) {
   console.log('error');
}
