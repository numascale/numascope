function rand() {
  return Math.random();
}

const ws = new WebSocket('ws://'+location.host+'/monitor');
var signedon = false;
var buttons;

function graph() {
   var time = new Date();

   var data = [{
     x: [time],
     y: [rand],
     mode: 'lines',
     line: {color: '#80CAF6'}
   }]

   Plotly.plot('graph', data);
}

function update() {
   var time = new Date();

   var update = {
      x:  [[time]],
      y: [[rand()]]
   }

   var olderTime = time.setMinutes(time.getMinutes() - 1);
   var futureTime = time.setMinutes(time.getMinutes() + 1);

   var minuteView = {
      xaxis: {
         type: 'date',
         range: [olderTime,futureTime]
      }
   };

   Plotly.relayout('graph', minuteView);
   Plotly.extendTraces('graph', update, [0])
}

function signon(data) {
   for (var i = 0; i < data["Tree"].length; i++) {
      for (var key in data["Tree"][i]) {
         if (!data["Tree"][i].hasOwnProperty(key))
            continue;

         buttons = document.createElement('details');
         var node = document.createElement('summary');
         buttons.appendChild(node);
         var text = document.createTextNode(key+' events');
         node.appendChild(text);

         elems = data["Tree"][i][key];

         for (var j in elems) {
            var btn = document.createElement('button')
            var text = document.createTextNode(elems[j]);
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
   var data = JSON.parse(e.data);

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

      graph();
      return;
   }

   console.log('recv: '+JSON.stringify(data, null, 2));
   update();
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
