// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

$(init);

function init() {
	WS.Connect();
	WS.AutoRegister();

	$("button").button();
	$(".buttonset").buttonset();

	$("#aspect").change(setAspect);
	setAspect();

	$("#CopyToView").click(   function() { copy("Preview", "View"); });
	$("#CopyToPreview").click(function() { copy("View", "Preview"); });
}

function copy(from, to) {
	from = "Settings."+from;
	to = "Settings."+to;
	var div = $("[sbContext='"+from+"']");
	var elems = div.find("[sbBind]");
	elems.each(function(idx, elem) {
		elem = $(elem);
		var f = from + '.' + elem.attr('sbBind');
		var t = to + '.' + elem.attr('sbBind');
		WS.Set(t, WS.state[f]);
	});
}

function setAspect() {
	var a = $("#aspect").val();
	$("iframe").removeClass("aspect4x3").removeClass("aspect16x9");
	$("iframe").addClass(a);
}
