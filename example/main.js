// Create WebSocket connection.
//const socket = new WebSocket('ws://localhost:8888/?token=abcd');
const socket = new WebSocket('ws://172.30.60.245/?token=abcd');

// Connection opened
socket.addEventListener('open', function (event) {
    //认证
    a ={type:0,payload:"no578999",channel:"channel1"}
    console.log(JSON.stringify(a))
    socket.send(JSON.stringify(a));
});

// Listen for messages
socket.addEventListener('message', function (event) {
    const e = document.getElementById("messages");
    e.innerHTML = e.innerHTML + 'Message from server '+ event.data +"<br/>";

    console.log('Message from server ', event.data)
});