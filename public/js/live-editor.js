jQuery(function($){
    const socket = io.connect()
    const $submit = $('#send')
    const $message = $('#message')
    const $chat = $('#chat')
    const $editor = $('#code')
    const $changeLang = $('#filelang')
    const $changeName = $('#filename')

    const doc = $editor.doc
    var lastChange = ""

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
                    socket.emit('langChange',$(this).val());
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
                    socket.emit('nameChange',$(this).val());
                }
            });
        }
    });
    
    editor.on("changes", function(cm,arr){
        if(arr[0].origin != undefined) {
            socket.emit("code:update", JSON.stringify(arr[0]))
        }
    });

    socket.on('code:update',function(data) {
        data = JSON.parse(data)
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
});