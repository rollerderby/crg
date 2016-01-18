// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

var Game = {
	List: function(callback) {
		$.getJSON('/JSON/Game/List', callback);
	},
	Adhoc: function(obj, callback, error) {
		$.ajax({
			type: "POST",
			url: '/JSON/Game/Adhoc',
			data: JSON.stringify(obj),
			success: callback,
			error: error,
			dataType: "json"
		});
	}
};
