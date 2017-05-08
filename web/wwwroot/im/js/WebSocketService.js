var WebSocketService = function( webSocket) {
	var webSocketService = this;
	
	var webSocket = webSocket;
	var TypeReq  = "req";
	var TypePush = "push";
    var TypeBreoatcast= "broatcast";
	
	this.hasConnection = false;
	
	this.welcomeHandler = function(json_obj) {

        webSocketService.hasConnection = true;
        console.log("welcomeHandler:",json_obj);

        webSocketService.subscripeGroup()
    };


    this.failedHandler = function(json_obj) {
        webSocketService.hasConnection = true;
        console.log("failedHandler:");
        console.log(json_obj);
        console.log("加入失败!")

    };

    this.subscripeGroupfailedHandler = function(json_obj) {
        console.log(json_obj);
        console.log("加入群组失败!")

    };

    this.subscripeGroupHandler = function(json_obj) {
        console.log(json_obj);
        console.log("加入群组成功!")

    };

    this.failedHandler = function(json_obj) {
        webSocketService.hasConnection = true;
        console.log("failedHandler:");
        console.log(json_obj);
        alert("加入失败!")

    };
	
	this.updateHandler = function(json_obj) {
		var newtp = false;
		//console.log( "updateHandler:" );
		 console.log( json_obj );
 
	}
	
	this.pushHandler = function(json_obj ) {
		console.log( "messageHandler:" );
        console.log( json_obj );
		data  = json_obj.data

        from_info = data.msg.from_info
        var from_sid = data.from_sid

        for(var i=0; i<GlobalContacts.length; i++)
        {
            if  (GlobalContacts[i].sid==from_sid){
                from_info = GlobalContacts[i];
            }
        }
        console.log( "from_info:" );
        console.log( from_info );

        obj = {
            username:from_info.username
            ,avatar: from_info.avatar
            ,id: from_info.id
            ,type: "friend"
			,mine:false
            ,content: data.msg.content
        }

        layui.use('layim', function(layim){
            layim.getMessage(obj);
        });
		
	}

    this.broatcastHandler = function( json_obj ) {
        console.log( "groupMessageHandler:" );
		data  = json_obj.data

        from_info = data.msg.from_info

        group_id = ""
        for(var i=0; i<GlobalGroups.length; i++)
        {
            if  (GlobalGroups[i].channel_id==data.group_channel_id){
                group_id = GlobalGroups[i].id;
            }
        }

        var obj = {
            username:from_info.username
            ,avatar: from_info.avatar
            ,id: group_id
            ,fromid:from_info.id
			,mine:false
            ,type: "group"
            ,content: data.msg.content
        }
        console.log( "messageGroupHandler obj:" );
        console.log( obj );

        layui.use('layim', function(layim){
            layim.getMessage(obj);
        });

    }
	
	this.closedHandler = function(json_obj) {

	}
	
	this.redirectHandler = function( json_obj ) {

		data = json_obj.data
		if (data.url) {
			if (authWindow) {
				authWindow.document.location = data.url;
			} else {
				document.location = data.url;
			}			
		}
	}
	
	this.noneHandler = function(json_obj) {
		 
	}
	
	this.processMessage = function( json_obj ) {
	    console.log("processMessage:");
        var fn
        if( json_obj.type=="response"){
             fn = webSocketService[json_obj.data.type + 'Handler'];
        }else{
             fn = webSocketService[json_obj.type + 'Handler'];
        }

		if (fn) {
			fn(json_obj);
		}
	}
	
	this.connectionClosed = function() {
		webSocketService.hasConnection = false;
		 
	};
	
	this.wrapReqMessage = function( _cmd,sid,reqid,msg ){
	    // { "header":{ "cmd":"", "seq_id":0,  "sid":"" , "token":"", "version":"1.0" ,"gzip":true}  , "type":"req", "data":{}  }
		var req_obj = {
            header: {
				cmd:_cmd,
				seq_id:reqid,
				sid:sid, 
				token:GlobalToken,
				version:"1.0",
				no_resp:false,
				gzip:false
			},
            type:TypeReq,
            data: msg,
        };
        console.log( req_obj );
		return  JSON.stringify(req_obj) 

	}

 
	this.wrapPushMessage = function( sid,msg ){
		//  { "header":{ "cmd":"", "seq_id":0,  "sid":"" , "token":"", "version":"1.0" ,"gzip":true}  , "type":"req", "data":{}  }
		var req_obj = {
            header: {
				cmd:"PushMessage",
				seq_id:0,
				sid:sid, 
				token:GlobalToken,
				version:"1.0",
				no_resp:true,
				gzip:false
			},
            type:TypePush,
            data: msg,
        };
        console.log( req_obj );
		return  JSON.stringify(req_obj)
	}
	 

	this.wrapPushGroupMessage = function( sid,msg ){
		//  { "header":{ "cmd":"", "seq_id":0,  "sid":"" , "token":"", "version":"1.0" ,"gzip":true}  , "type":"req", "data":{}  }
		var req_obj = {
            header: {
				cmd:"PushGroupMessage",
				seq_id:0,
				sid:sid, 
				token:GlobalToken,
				version:"1.0",
				no_resp:true,
				gzip:false
			},
            type:TypeBreoatcast,
            data: msg,
        };
        console.log( req_obj );
        return  JSON.stringify(req_obj)
	}

	this.sendMessage = function( sid, msg  ) {
		var sendObj = {
			type: 'message',
			message: msg,
			id:sid
		};
        str = this.wrapReqMessage( 'Message',sid,0,sendObj)
		webSocket.send(str);
	}
	 

	this.pushMessage = function( sid, msg  ) {
		console.log("pushMessage:");
        console.log( sid );
        console.log( msg );
		str = this.wrapPushMessage( sid,msg)
		webSocket.send(str);
	}
	this.pushGroupMessage = function( sid, msg  ) {

		str = this.wrapPushGroupMessage( sid,msg)
		webSocket.send(str);
	}


	this.authorize = function(token,sid) {
		var sendObj = {
			type: 'authorize',
			token: token,
			sid: sid
		};
        str = this.wrapReqMessage( 'Authorize',sid,0,sendObj)
		webSocket.send(str);
	}

    this.subscripeGroup = function( ) {
        var sendObj = {
            type: 'SubscripeGroup',
            token: GlobalToken,
            sid: GlobalSid
        };
        str = webSocketService.wrapReqMessage( 'SubscripeGroup',GlobalSid,0,sendObj)
        webSocket.send(str);
    }

}