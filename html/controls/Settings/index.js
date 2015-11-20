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
}

function setAspect() {
	var a = $("#aspect").val();
	$("iframe").removeClass("aspect4x3").removeClass("aspect16x9");
	$("iframe").addClass(a);
}
