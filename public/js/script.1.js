var util = util || {};
util.toArray = function(list) {
  return Array.prototype.slice.call(list || [], 0);
};
htmlToText = function(str) {
  return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

var Terminal = Terminal || function(cmdLineContainer, outputContainer) {
  window.URL = window.URL || window.webkitURL;
  window.requestFileSystem = window.requestFileSystem || window.webkitRequestFileSystem;

  var cmdLine_ = document.querySelector(cmdLineContainer);
  var output_ = document.querySelector(outputContainer);

  const CMDS_ = [
    'cat', 'clear', 'save', 'echo', 'help', 'open', 'whoami',
    'expand', 'minimize', 'shrink', 'rm', 'cd', 'rmdir', 'dir'
  ];
  
  var fs_ = null;
  var cwd_ = null;
  var terminal = $(outputContainer.split(" ")[0])
  var history_ = [];
  var histpos_ = 0;
  var histtemp_ = 0;
  
  loadHistory()
  
  terminal.on('click', function(e) {
    cmdLine_.focus();
  }, false);

  cmdLine_.addEventListener('click', inputTextClick_, false);
  cmdLine_.addEventListener('keydown', historyHandler_, false);
  cmdLine_.addEventListener('keydown', processNewCommand_, false);
  function inputTextClick_(e) {
    this.value = this.value;
  }
  
  function saveHistory() {
    var d = new Date();
    d.setTime(d.getTime() + 60*60*1000);
    var expires = "expires="+ d.toUTCString();
    data = JSON.stringify(history_)
    document.cookie = "terminal_history=" + data + ";" + expires + ";path=/";
  }
  
  function loadHistory() {
    var name = "terminal_history=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for(var i = 0; i <ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            history_ = JSON.parse(c.substring(name.length, c.length))
            histpos_ = history_.length
        }
    }
  }

  //
  function historyHandler_(e) {
    if (history_.length) {
      if (e.keyCode == 38 || e.keyCode == 40) {
        if (history_[histpos_]) {
          history_[histpos_] = this.value;
        } else {
          histtemp_ = this.value;
        }
      }

      if (e.keyCode == 38) { // up
        histpos_--;
        if (histpos_ < 0) {
          histpos_ = 0;
        }
      } else if (e.keyCode == 40) { // down
        histpos_++;
        if (histpos_ > history_.length) {
          histpos_ = history_.length;
        }
      }

      if (e.keyCode == 38 || e.keyCode == 40) {
        this.value = history_[histpos_] ? history_[histpos_] : histtemp_;
        this.value = this.value; // Sets cursor to end of input.
      }
    }
  }

  //
  function processNewCommand_(e) {

    if (e.keyCode == 9) { // tab
      e.preventDefault();
      // Implement tab suggest.
    } else if (e.keyCode == 13) { // enter
      this.readOnly = true
      // Save shell history.
      if (this.value) {
        history_[history_.length] = this.value;
        histpos_ = history_.length;
        saveHistory()
      }

      // Duplicate current input and append to output section.
      var line = this.parentNode.parentNode.cloneNode(true);
      line.removeAttribute('id')
      line.classList.add('line');
      var input = line.querySelector('input.cmdline');
      input.autofocus = false;
      input.readOnly = true;
      output_.appendChild(line);

      if (this.value && this.value.trim()) {
        var args = this.value.split(' ').filter(function(val, i) {
          return val;
        });
        var cmd = args[0].toLowerCase();
        args = args.splice(1); // Remove cmd from arg list.
      }

      switch (cmd) {
        case 'cat':
          var url = args.join(' ');
          if (!url) {
            output('Usage: ' + cmd + ' https://s.codepen.io/...');
            output('Example: ' + cmd + ' https://s.codepen.io/AndrewBarfield/pen/LEbPJx.js');
            break;
          }
          $.ajax({
              url: url,
              type: 'GET',
              crossDomain: true,
              headers:{"Access-Control-Allow-Credentials":"true","Access-Control-Allow-Origin":"true"},
              success: function(data) {
                output('<pre style="height:150px;overflow:scroll;">' + htmlToText(data) + '</pre>');
              },
              error: function() {
                output('<p>Could retrieve data at '+url+'</p>')
              }
          })         
          break;
        case 'clear':
          output_.innerHTML = '';
          this.value = '';
          return;
        case 'date':
          output( new Date() );
          break;
        case 'echo':
          output( args.join(' ') );
          break;
        case 'help':
          output('<div class="ls-files">' + CMDS_.join('<br>') + '</div>');
          break;
        case 'uname':
          output(navigator.appVersion);
          break;
        case 'whoami':
          output("you")
          break;
        default:
          if (cmd) {
            output(cmd + ': command not found');
          }
      };

      terminal.animate({ scrollTop: terminal[0].scrollHeight }, "slow");
    }
  }

  //
  function formatColumns_(entries) {
    var maxName = entries[0].name;
    util.toArray(entries).forEach(function(entry, i) {
      if (entry.name.length > maxName.length) {
        maxName = entry.name;
      }
    });

    var height = entries.length <= 3 ?
        'height: ' + (entries.length * 15) + 'px;' : '';

    // 12px monospace font yields ~7px screen width.
    var colWidth = maxName.length * 7;

    return ['<div class="ls-files" style="-webkit-column-width:',
            colWidth, 'px;', height, '">'];
  }

  //
  function output(html) {
    output_.insertAdjacentHTML('beforeend', '<p>' + html + '</p>');
    this.value = ''; // Clear/setup line for next input.
  }

  // Cross-browser impl to get document's height.
  function getDocHeight_() {
    var d = document;
    return Math.max(
        Math.max(d.body.scrollHeight, d.documentElement.scrollHeight),
        Math.max(d.body.offsetHeight, d.documentElement.offsetHeight),
        Math.max(d.body.clientHeight, d.documentElement.clientHeight)
    );
  }

  //
  return {
    init: function() {
      output('<img align="left" src="/assets/images/netcode.png" width="100" height="100" style="padding: 10px 10px 10px 10px"><h2 style="letter-spacing: 4px"><b>NetCode Web Terminal</b></h2><p>' + new Date() + '</p><p>Enter "help" for more information.</p>');
    },
    output: output
  }
}

$('.prompt').html('[user@HTML5] >');

// Initialize a new terminal object
var term = new Terminal('#input-line .cmdline', '#outer output');
term.init();
let search = function(list,vals,func) {
	if(vals.length > 0) {
		let v = 0
		let i = list.indexOf(vals[v++])
		let f = i
		if(i++ >= 0) {
			for(;v<vals.length && i<list.length;) {
				console.log(i+ " " +v)
				if(list[i++] == vals[v++]) {
					continue
				}
			}
			console.log(i+ " " +v)
			if(vals.length == v) {
				func(vals,f)
				return true
			}
		}
		return false
	}
	return false
}

var netconsole = $("#netconsole")
netconsole.keypress(function(e) {
    if(e.which == 13) {
    	let tokens = netconsole.val().split(" ")
        search(tokens,["room","join"], function(vals,f) {
        	let i = vals.length
        	console.log(i + " " + f + " " + tokens.length)
        	if(tokens.length > i + f + 1) {
        		console.log("Joined room " + tokens[i])
        	} else {
        		if(!search(tokens,vals.push("as","default","admin"), (tokens,vals,f) => {
      				console.log("Joined as admin")
      			})) console.log("No room specified")
        	}
      		
        } )
    }
});

