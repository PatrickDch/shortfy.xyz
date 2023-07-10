function makeid(length) {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  let counter = 0;
  while (counter < length) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
    counter += 1;
  }
  return result;
}
  
let regid = makeid(5)
console.log("RegID "+regid);

function short() {

  document.getElementById('gifimg').src = "";
  document.getElementById("gif").style.visibility = "hidden";
  document.getElementById('result').href = "";
  document.getElementById("url").style.visibility = "hidden";
  document.getElementById('result').innerHTML = "";
  document.getElementById("helpurl").style.visibility = "hidden";

    if (document.getElementById("website").value === "" && !document.getElementById("website").value && !document.getElementById("website").value) {
    newebsite = "https://shortfy.xyz";
    console.log("empty URL");
  
    const endpointbs = 'https://shortfy.xyz/bs';
 
    let data = {
      RegID: regid
    }
  
    let fetchData = {
      method: 'POST',
      body: JSON.stringify(data),
      headers: new Headers({
        'Content-Type': 'application/json; charset=UTF-8'
      })
    }
    fetch(endpointbs, fetchData)
    .then(function(response) { return response.json(); })
    .then(function(json) {
    console.log(json)
    var gifurl = json.url;
    $("#gif").prependTo("#active");
    document.getElementById('gifimg').src = gifurl;
    document.getElementById("gif").style.visibility = "visible";
    var gifurl = ""
})   
$("#url").appendTo("#gif");
  document.getElementById('result').href = newebsite;
  document.getElementById("url").style.visibility = "visible";
  document.getElementById('result').innerHTML = newebsite;
  navigator.clipboard.writeText(newebsite);
  Toastify({
    text: "Trying to fix invalid URL. Enter your URL",
    gravity: "top",
    position: "center",
    style: {
        background: "linear-gradient(to right, #d50000, #ecce37)",
    }
}).showToast();
throw new Error("URL Broken");
}
 else {
    
    var website = document.getElementById("website").value;

        console.log("forward");
        console.log(website);
        if (!website.includes("tps:") && !website.includes("tp:") && !website.includes("net:")) {
          console.log("guessing protocoll")

          if (website.includes(":")) {
          var website = "https://"+website.substring(website.indexOf(":") + 1);
          }
          if (!website.includes(":")) {
          var website = "https://"+website
          }

          var urled = new URL(website);

          if  (!urled.protocol == "https:" && !urled.protocol == "http:" && !urled.protocol == "magnet:" && !urled.protocol == "xml:") {
            website = "https://developer.mozilla.org/en-US/docs/Learn/Common_questions/Web_mechanics/What_is_a_URL";
            console.log("wrong url");
                   Toastify({
                       text: "Invalid URL protocol. Replaced URL",
                       gravity: "top",
                       position: "center",
                       style: {
                           background: "linear-gradient(to right, #d50000, #ff1744)",
                       }
       
                   }).showToast();
                   Toastify({
               text: "Copied to clipboard!",
               gravity: "top",
               position: "center"
             }).showToast();
             $("#helpurl").prependTo("#active");
                   document.getElementById("helpurl").style.visibility = "visible";
                   navigator.clipboard.writeText(website);
                   console.log(website);
                   return
               }}
         
        }

        if (website.startsWith('magnet:') && website.startsWith('xml:')) {
          var posturl = website;
        }
        else {
        var posturl = new URL(website);
        }
        console.log(posturl);

        const endpoint = 'https://shortfy.xyz/p/create/';

        let data = {
          url: posturl
        }
        
        let fetchData = {
          method: 'POST',
          body: JSON.stringify(data),
          headers: new Headers({
            'Content-Type': 'application/json; charset=UTF-8'
          })
        }
        
       fetch(endpoint, fetchData)
        .then(function(response) { return response.json(); })
        .then(function(json) {
            console.log(json)
            
            if (json.status == "400")
            {
        
              const endpointbs = 'https://shortfy.xyz/bs';
 
              let data = {
                RegID: regid
              }
            
              let fetchData = {
                method: 'POST',
                body: JSON.stringify(data),
                headers: new Headers({
                  'Content-Type': 'application/json; charset=UTF-8'
                })
              }
              fetch(endpointbs, fetchData)
              .then(function(response) { return response.json(); })
              .then(function(json) {
              console.log(json)
              var gifurl = json.url;
              $("#gif").prependTo("#active");
              $("#url").appendTo("#gif");
              document.getElementById('gifimg').src = gifurl;
              document.getElementById("gif").style.visibility = "visible";
              var gifurl = ""
          })   
             var url = "is broken..."
            }
            else {
            var url = json.url;
            $("#url").prependTo("#active");
            }
      
      let date = moment().add(1, "d");
			document.getElementById('result').href = url;
      document.getElementById('valid').innerHTML = date;
			document.getElementById('result').innerHTML = url;
      document.getElementById("url").style.visibility = "visible";
			navigator.clipboard.writeText(url);
			Toastify({
				text: "Copied to clipboard!",
				gravity: "top",
				position: "center"
			}).showToast();
        });
      
}