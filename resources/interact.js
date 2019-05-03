function rand() {
  return Math.random();
}

var time = new Date();

var data = [{
  x: [time],
  y: [rand],
  mode: 'lines',
  line: {color: '#80CAF6'}
}]

Plotly.plot('graph', data);

var cnt = 0;

var interval = setInterval(function() {
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

  if(cnt === 100) clearInterval(interval);
}, 1000);

const ws = new WebSocket('ws://'+location.host+'/monitor');
var signedon = false;
var xyz;

function signon(data) {
xyz=data;

   for (var i = 0; i < data["Tree"].length; i++) {
      for (var key in data["Tree"][i]) {
         if (!data["Tree"][i].hasOwnProperty(key))
            continue;

         let elem = document.createElement('details');
         var node = document.createElement('summary');
         elem.appendChild(node);
         var text = document.createTextNode(key);
         node.appendChild(text);

         elems = data["Tree"][i][key];
         for (var j in elems) {
            var para = document.createElement("p");
            var text = document.createTextNode(elems[j]);
            para.appendChild(text);
            elem.appendChild(para);
         }

         let container = document.querySelector("#events");
         container.appendChild(elem);
      }
   }

   signon = true;
}

ws.onmessage = function(e) {
//   console.log('recv '+e.data);

   if (signedon == false) {
      signon(JSON.parse(e.data))
   }
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
