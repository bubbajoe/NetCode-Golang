jQuery(function($) {
    const socket = io.connect()
    const $submit = $('#send')
    const $message = $('#message')
    const $chat = $('#chat')
    const $editor = $('#code')
    const $changeLang = $('#filelang')
    const $changeName = $('#filename')
    
    const doc = $editor.doc
    var lastChange = ""

    function getCookieValue(a) {
		var b = document.cookie.match('(^|;)\\s*' + a + '\\s*=\\s*([^;]+)');
		return b ? b.pop() : '';
	}

    // Send and append message
    $submit.click(function() {
        var msg = $message.val().trim(); // removes excess white space from string
        if(msg != '') {
            socket.emit('send',msg); // send to the socket server
            $chat.append("<div id='me'>"+msg+"<br/><div>"); // adds my message 
            $chat.scrollTop($chat.height());
        }
        $message.val('');
        $message.select();
    });

    $message.click(function(e) {

    });

    // Enter to send
    $message.keypress(function(e){
        if(e.which == 13) { // 13 = enter key
            $submit.click();
            return false;
        }
    });

    $changeLang.click(function(e) {
        if($('#changeLangInput').length == 0) {
            var filename = $(this).html();
            $(this).html('<input type="textbox" style="width: 200px" id="changeLangInput"></input>');
            $('#changeLangInput').val(filename);
            $('#changeLangInput').focus();
            $('#changeLangInput').select();
            $('#changeLangInput').keypress(function(e){
                //$(this).css("width",$(this).val().length*2);
                if(e.which == 13) { // 13 = enter key
                    $changeLang.html($(this).val());
                    socket.emit('langChange',JSON.stringify({sessionID:getCookieValue('session'),data:$(this).val()}));
                }
            });
        }
    });

    $changeName.click(function(e) {
        if($('#changeNameInput').length == 0) {
            var filename = $(this).html();
            $(this).html('<input type="textbox" style="width: 200px" id="changeNameInput"></input>');
            $('#changeNameInput').val(filename);
            $('#changeNameInput').focus();
            $('#changeNameInput').select();

            $('#changeNameInput').keypress(function(e){
                //$(this).css("width",$(this).val().length*2);
                if(e.which == 13) { // 13 = enter key
                    $changeName.html($(this).val());
                    socket.emit('nameChange',JSON.stringify({sessionID:getCookieValue('session'),data:$(this).val()}));
                }
            });
        }
    });
    
    editor.on("changes", function(cm,arr){
        if(arr[0].origin != undefined) {
            socket.emit("code:update", JSON.stringify({data:arr[0],sessionID:getCookieValue('session')}))
        }
    });

    socket.on('code:update',function(pkt) {
        pkt = JSON.parse(pkt)
        let data = pkt.data
        let i = 0;
        if(data.text.length > 0 || data.text[0] != '')
            data.text.forEach(function(item,index) {
                data.from.line += index;
                data.to.line += index;
                editor.doc.replaceRange(
                    item+((index != data.text.length-1) ? "\n" : ""),
                    data.from,
                    data.to
                );
                data.from.line -= index;
                data.to.line -= index;
            }, this);
        else
            data.removed.forEach(function(item,index) {
                editor.doc.replaceRange(
                    item+((index != data.removed.length-1)? "\n" : ""),
                    data.from,
                    data.to
                );
            }, this);
    });

    socket.on('message',function(data) {
        if(data.trim() != '') {
            $chat.append("<div id='other'>"+data+"<br/><div>"); // adds new message
            $chat.scrollTop($chat.height()); // Scrolls down to the bottom
        }
    });
    
    var util = util || {};
    util.toArray = function(list) {
      return Array.prototype.slice.call(list || [], 0);
    };
    
    var htmlToText = (str) => {
      return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }
    
    var getCookie = (cname) => {
        var name = cname + "=";
        var decodedCookie = decodeURIComponent(document.cookie);
        var ca = decodedCookie.split(';');
        for(var i = 0; i <ca.length; i++) {
            var c = ca[i];
            while (c.charAt(0) == ' ') {
                c = c.substring(1);
            }
            if (c.indexOf(name) == 0) {
                return c.substring(name.length, c.length);
            }
        }
        return "";
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
      var line = ""
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
        document.cookie = "terminal_history=" + data + ";" + expires + ";";
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
    	        try {
                	history_ = JSON.parse(c.substring(name.length, c.length))
            	} catch(e) {
            		history_ = {}
            	}
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
    
        if (e.keyCode == 9) {
          e.preventDefault();
        } else if (e.keyCode == 13) {
          if (this.value) {
            history_[history_.length] = this.value;
            histpos_ = history_.length;
            if(history_.length >= 20) {
              history_.shift()
              histpos_--
            }
            saveHistory()
          }
          
          line = this.parentNode.parentNode.cloneNode(true)
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
            this.value = ""
          } else return
          
          socket.emit("terminal:command",cmd);
        }
      }
      
    socket.on('terminal:responce',function(res) {
        output(res)
    });
    
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
        terminal.animate({ scrollTop: terminal[0].scrollHeight }, 800);
        this.value = ""; // Clear/setup line for next input.
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
    
    $('.prompt').html('username: $ ');
    
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
});