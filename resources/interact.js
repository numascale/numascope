function rand() {
  return Math.random();
}

const ws = new WebSocket('ws://'+location.host+'/monitor');
let signedon = false;
let buttons;

function graph(headings) {
   let data = [];

   for (const heading of headings) {
      data.push({
         name: heading,
         mode: 'lines',
         x: [0],
         y: [0]
      });
   }

   const layout = {
      yaxis: {
         title: 'events'
      }
   }

   Plotly.react('graph', data, layout);
}

function update(data) {
   let time = new Date();

   let update = {
      x: [[time]],
      y: [[rand()]]
      }

   let olderTime = time.setMinutes(time.getMinutes() - 1);
   let futureTime = time.setMinutes(time.getMinutes() + 1);

   let minuteView = {
      xaxis: {
         type: 'date',
         range: [olderTime,futureTime]
      }
   };

   Plotly.relayout('graph', minuteView);
   Plotly.extendTraces('graph', update, [0])
}

function signon(data) {
   for (let i = 0; i < data["Tree"].length; i++) {
      for (const key in data["Tree"][i]) {
         if (!data["Tree"][i].hasOwnProperty(key))
            continue;

         buttons = document.createElement('details');
         let node = document.createElement('summary');
         buttons.appendChild(node);
         let text = document.createTextNode(key+' events');
         node.appendChild(text);

         elems = data["Tree"][i][key];

         for (const elem of elems) {
            let btn = document.createElement('button')
            let text = document.createTextNode(elem);
            btn.appendChild(text);
            btn.className = 'btn btn-light btn-sm m-1';
            buttons.appendChild(btn);
         }

         let container = document.querySelector("#events");
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

   // handle enabled updates
   if (data[0] == 'enabled') {
      // drop 'enabled' element
      data.shift();

      for (let btn of buttons.childNodes) {
         if (!btn.className.startsWith('btn')) {
            continue;
         }

         btn.className = data.includes(btn.firstChild.nodeValue) ? 'btn btn-primary btn-sm m-1' : 'btn btn-light btn-sm m-1';
      }

      graph(data);
      return;
   }

   console.log('recv: '+JSON.stringify(data, null, 2));
   update(data);
}

ws.onopen = function(e) {
   ws.send('463ba1974b06')
   console.log('authenticated');
}

ws.onclose = function(e) {
   console.log('closed');
}

ws.onerror = function(e) {
   console.log('error');
}
