// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

/* This file requires base jQuery; other required jQuery plugins are automatically included below. */
if (typeof $ == "undefined") {
	alert("You MUST include jQuery before this file!");
	throw("You MUST include jQuery before this file!");
}

function _includeUrl(url) {
	var filename = url.replace(/^.*[\/]/g, "");
	/* Use HTTP HEAD to verify url exists before adding it to the document */
	if ($.ajax(url, { async: false, type: "HEAD", global: false }).status != 200)
		return;
	if (/\.[cC][sS][sS](\?.*)?$/.test(url) && !$("head link[href='"+url+"'],head link[href='"+filename+"']").length)
		$("<link>").attr({ href: url, type: "text/css", rel: "stylesheet"}).appendTo("head");
	else if (/\.[jJ][sS](\?.*)?$/.test(url) && !$("head script[src='"+url+"'],head script[src='"+filename+"']").length)
		$("<script>").attr({ src: url, type: "text/javascript" }).appendTo("head");
}

function _include(dir, files) {
	if (!files) {
		files = dir;
		dir = undefined;
	}
	if (!$.isArray(files))
		files = [ files ];
	$.each(files, function() { _includeUrl((dir?dir+"/":"")+this); });
}

_include("/external/jquery-ui", [ "jquery-ui.js", "css/default/jquery-ui.css" ]);

_include("/external/jquery-plugins/isjquery/jquery.isjquery.js");
_include("/external/jquery-plugins/string/jquery.string.js");

/* Good places to find fonts are:
 * http://fontspace.com/
 * http://fontsquirrel.com/
 * Also very handy is the @font-face generator at fontsquirrel:
 * http://www.fontsquirrel.com/fontface/generator
 */
_include("/fonts", [
	"liberationsans/stylesheet.css", "roboto/stylesheet.css" ]);

/* Core functionality */
_include("/javascript", [
	"WS.js",
	"time.js",
	"windowfunctions.js",
	"autofit.js",
]);
/*
_include("/json", [
	"Rulesets.js",
	"Game.js",
]);
*/

$(function() {
	if (/\.html$/.test(window.location.pathname)) {
		_include(window.location.pathname.replace(/\.html$/, ".css"));
		_include(window.location.pathname.replace(/\.html$/, ".js"));
	} else if (/\/$/.test(window.location.pathname)) {
		_include(window.location.pathname + "index.css");
		_include(window.location.pathname + "index.js");
	}
});

$.fn.filterByData = function(prop, val) {
	return this.filter(
		function() { return $(this).data(prop)==val; }
	);
}

function isTrue(value) {
	if (typeof value == "boolean")
		return value;
	else
		return (String(value).toLowerCase() == "true");
}
