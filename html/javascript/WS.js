// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

var WS = {

	connectCallback: null,
	connectTimeout: null,
	callbacks: new Array(),
	Connected: false,
	state: { },
	debug: false,

	Connect: function(callback) {
		WS.connectCallback = callback;
		WS._connect();
	},

	_connect: function() {
		WS.connectTimeout = null;
		var url = (document.location.protocol == "http:" ? "ws" : "wss") + "://";
		url += document.location.host + "/ws";

		if(WS.Connected != true || !WS.socket) {
			if(WS.debug) console.log("WS", "Connecting the websocket at " + url);

			WS.socket = new WebSocket(url);
			WS.socket.onopen = function(e) {
				WS.Connected = true;
				if(WS.debug) console.log("WS", "Websocket: Open");
				$(".ConnectionError").addClass("Connected");
				req = {
					action: "Register",
					data: new Array()
				};
				$.each(Object.keys(WS.state), function(idx, k) {
					WS.triggerCallback(k, null);
				});
				WS.state = {};
				$.each(WS.callbacks, function (idx, c) {
					req.data.push(c.path);
				});
				if (req.data.length > 0) {
					WS.send(JSON.stringify(req));
				}
				if (WS.connectCallback != null)
					WS.connectCallback();
			};
			WS.socket.onmessage = function(e) {
				json = JSON.parse(e.data);
				// console.log(e.data, json);
				if (WS.debug) console.log("WS", json);
				if (json.authorization != null)
					alert(json.authorization);
				if (json.state != null)
					WS.processUpdate(json.state);
			};
			WS.socket.onclose = function(e) {
				WS.Connected = false;
				console.log("WS", "Websocket: Close", e);
				$(".ConnectionError").removeClass("Connected");
				if (WS.connectTimeout == null)
					WS.connectTimeout = setTimeout(WS._connect, 1000);
			};
			WS.socket.onerror = function(e) {
				console.log("WS", "Websocket: Error", e);
				$(".ConnectionError").removeClass("Connected");
				if (WS.connectTimeout == null)
					WS.connectTimeout = setTimeout(WS._connect, 1000);
			}
		} else {
			// better run the callback on post connect if we didn't need to connect
			if(WS.connectCallback != null) WS.connectCallback();
		}
	},

	send: function(data) {
		if (WS.socket != null && WS.socket.readyState === 1) {
			WS.socket.send(data);
		}
	},

	Command: function(command, data) {
		if (!Array.isArray(data))
			data = [data];
		req = {
			action: command,
			data: data
		};
		WS.send(JSON.stringify(req));
	},

	NewObject: function(name, data) {
		req = {
			action: "NewObject",
			field: name,
			fieldData: {}
		};
		for (var prop in data) {
			req.fieldData[prop] = data[prop].toString();
		}
		WS.send(JSON.stringify(req));
	},

	Set: function(key, value) {
		WS.Command("Set", [key, value]);
	},

	triggerCallback: function (k, v) {
		var callbackCalled = false;
		for (idx = 0; idx < WS.callbacks.length; idx++) {
			c = WS.callbacks[idx];
			if (c.callback == null)
				continue;
			if (patternMatch(k, c.path)) { // k.indexOf(c.path) == 0) {
				try {
					c.callback(k, v);
					callbackCalled = true;
				} catch (err) {
					console.log(err.message, err.stack);
				}
			}
		}
		if (!callbackCalled) {
			console.log("No callback found for " + k);
		}
	},

	processUpdate: function (state) {
		for (var prop in state) {
			// update all incoming properties before triggers
			// dependency issues causing problems
			// console.log(prop, state[prop]);
			WS.state[prop] = state[prop];
		}

		for (var prop in state) {
			if (state[prop] == null)
				WS.triggerCallback(prop, state[prop]);
		}
		for (var prop in state) {
			if (state[prop] != null)
				WS.triggerCallback(prop, state[prop]);
		}
	},

	Register: function(paths, options) {
		if ($.isFunction(options))
			options = { triggerFunc: options };

		var callback = null;
		if (options == null) {
			callback = null;
		} else {
			if (options.triggerFunc != null) {
				callback = options.triggerFunc;
			} else {
				var elem = options.element;
				if (options.css != null) {
					callback = function(k, v) { elem.css(options.css, v); };
				} else if (options.attr != null) {
					callback = function(k, v) { elem.attr(options.attr, v); };
				} else {
					if (elem.hasClass("AutoFit")) {
						elem.empty();
						var div = $("<div>").css("width", "100%").css("height", "100%").appendTo(elem);
						elem = $("<a>").appendTo(div);
						var autofit = _autoFit.enableAutoFitText(div);

						callback = function(k, v) {
							elem.text(v);
							if (elem.data("lastText") != v) {
								elem.data("lastText", v);
								autofit();
							}
						};
					} else if (elem.parent().hasClass("AutoFit")) {
						var autofit = _autoFit.enableAutoFitText(elem.parent());
						callback = function(k, v) {
							elem.text(v);
							if (elem.data("lastText") != v) {
								elem.data("lastText", v);
								autofit();
							}
						};
					} else {
						callback = function(k, v) { elem.text(v); };
					}
				}
			}

			if (options.modifyFunc != null) {
				var origCallback = callback;
				callback = function(k, v) { origCallback(k, options.modifyFunc(k, v)); };
			}
		}

		if (!$.isArray(paths)) {
			paths = [ paths ];
		}

		$.each(paths, function(idx, path) {
			WS.callbacks.push( { path: path, callback: callback } );
		});

		req = {
			action: "Register",
			data: paths
		};
		WS.send(JSON.stringify(req));
	},

	Unregister: function(paths, func) {
		var callbacks = null;
		req = {
			action: "Register",
			data: new Array()
		};
		$.each(paths, function(idx, path) {
			$.each(WS.callbacks, function(idx, callback) {
				var foundNonMatch = false;
				if (callback.path == path) {
					if (callback.callback == func) {
						if (callbacks == null)
							callbacks = WS.callbacks;
						callbacks = callbacks.splice(idx, 1);
					} else {
						foundNonMatch = true;
					}
				}
				if (!foundNonMatch)
					req.data.push(path);
			});
		});
		if (callbacks != null) {
			WS.callbacks = callbacks;
			WS.send(JSON.stringify(req));
		}
	},

	getPaths: function(elem, attr) {
		var list = elem.attr(attr).split(",");
		var path = WS._getContext(elem);
		paths = new Array();
		$.each(list, function(idx, item) {
			item = $.trim(item);
			if (path != null)
				item = path + '.' + item;
			paths.push(item);
		});
		return paths;
	},

	AutoRegister: function() {
		$.each($("[sbCopyDiv]"), function(idx, div) {
			div = $(div);
			div.html($(div.attr("sbCopyDiv")).html());
		});
		$.each($("[sbDisplay]"), function(idx, elem) {
			elem = $(elem);
			var paths = WS.getPaths(elem, "sbDisplay");
			if (paths.length > 0)
				WS.Register(paths, { element: elem, modifyFunc: window[elem.attr("sbModify")] });
		});
		$.each($("[sbTrigger]"), function(idx, elem) {
			elem = $(elem);
			var sbTrigger = window[elem.attr("sbTrigger")];
			if (sbTrigger == null)
				return;

			var paths = WS.getPaths(elem, "sbTriggerOn");
			if (paths.length > 0)
				WS.Register(paths, { triggerFunc: sbTrigger } );
		});
		$.each($("[sbCommand]"), function(idx, elem) {
			elem = $(elem);
			var cmd = elem.attr("sbCommand");
			var field = elem.attr("sbCommandField");
			if (field != null) {
				cmd = WS._getContext(elem, "sbCommandField") + "." + cmd;
			} else {
				cmd = WS._getContext(elem) + "." + cmd;
			}
			var dataAttr = elem.attr("sbData");
			var data = new Array();
			if (dataAttr != null) {
				data = dataAttr.split(" ");
			}
			elem.click(function() {
				var d = data.slice(0);
				if (elem.val() != null)
					d.push(elem.val());
				WS.Command(cmd, d);
			});
		});
		$.each($("[sbBind]"), function(idx, elem) {
			elem = $(elem);
			var field = WS._getContext(elem, "sbBind");

			if (elem.is("button")) {
				var data = elem.attr("sbData");
				if (data == null)
					return;
				elem.click(function(ev) {
					WS.Set(field, data);
				});
				WS.Register(field, function(k, v) {
					elem.toggleClass("active", v == data);
				});
			} else {
				elem.click(function(ev) {
					WS.Set(field, elem.val());
				});
				WS.Register(field, function(k, v) {
					if (k == field) {
						elem.val(v);
					}
				});
			}
		});
	},

	_getContext: function(elem, attr) {
		if (attr == null)
			attr = "sbContext";

		var parent = elem.parent();
		var ret = '';
		if (parent.length > 0)
			ret = WS._getContext(parent, "sbContext");
		var context = elem.attr(attr);
		if (context != null)
			ret = (ret != '' ? ret + '.' : '') + context;
		return ret;
	},
};

function patternMatch(value, pattern) {
	// Special case, if pattern is empty, it matches
	if (pattern == "") {
		return true;
	}
	while (true) {
		var id = pattern.indexOf("(*)");
		if (id == -1) {
			break;
		}
		id = id + 1;
		// check if value is long enough
		if (value.length < id) {
			return false;
		}

		// check everything leading up to the open paren
		if (value.substring(0, id) != pattern.substring(0, id)) {
			return false;
		}

		// prefix matched, now look for the close paren
		var rparen = value.indexOf(")");
		if (rparen == -1) {
			return false;
		}

		value = value.substring(rparen+1);
		pattern = pattern.substring(id+2);
	}

	if (value == pattern) {
		return true;
	}
	if (value.indexOf(pattern+".") == 0) {
		return true;
	}

	// match if pattern is empty and value is empty or starts with a dot
	if (pattern.length == 0) {
		var ret = value.length == 0 || value[0] == '.';
		return ret;
	}

	// look for trailing *
	if (pattern.substring(pattern.length-1) == "*") {
		var ret = value.indexOf(pattern.substring(0, pattern.length - 1)) == 0;
		return ret;
	}

	return value == pattern;
}
