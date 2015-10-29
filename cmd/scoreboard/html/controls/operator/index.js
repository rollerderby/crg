$(init);

function init() {
	// $("div").each(function (idx, elem) {
	// 	elem = $(elem);
	// 	if (elem.children().length == 0) {
	// 		elem.append($("<span>").text(elem.attr("class")));
	// 	}
	// });

	$(window).resize(function () {
		width = $(window).width();
		height = $(window).height();
		console.log("RESIZE", width, height);
	});
	width = $(window).width();
	height = $(window).height();
	tWidth = window.outerWidth;
	tHeight = window.outerHeight;
	offsetWidth = tWidth - width;
	offsetHeight = tHeight - height;
	console.log(width, height, tWidth, tHeight, offsetWidth, offsetHeight);
	window.resizeTo(1366 + offsetWidth, 768 + offsetHeight);
	

	$("[sbCopyDiv]").each(function(idx, div) {
		div = $(div);
		div.html($(div.attr("sbCopyDiv")).html());
	});
	$("button").button();
	$(".buttonset").buttonset();
	// $("span").addClass("ui-widget");
	
	WS.Connect();
	WS.AutoRegister();

	$(["Period", "Jam", "Lineup", "Timeout", "Intermission"]).each(function(idx, clock) {
		registerButtonCommand(".Clock." + clock + " .Time .Down button", "ClockAdjustTime", [clock, "-1000"]);
		registerButtonCommand(".Clock." + clock + " .Time .Up button", "ClockAdjustTime", [clock, "1000"]);
		registerButtonCommand(".Clock." + clock + " .Num .Down button", "ClockAdjustNumber", [clock, "-1"]);
		registerButtonCommand(".Clock." + clock + " .Num .Up button", "ClockAdjustNumber", [clock, "1"]);
	});

	registerButtonCommand(".MasterControls .StartJam", "StartJam");
	registerButtonCommand(".MasterControls .StopJam", "StopJam");
	registerButtonCommand(".MasterControls .Timeout", "Timeout");

	$(".startDebug").click(function() {  WS.Register(["ScoreBoard.State"], debugCallback); });
	$(".stopDebug").click(function() { WS.Unregister(["ScoreBoard.State"], debugCallback); });
}

function registerButtonCommand(select, command, data) {
	$(select).click(function() { console.log("COMMAND: ", command, data); WS.Command(command, data); });
}

function startClock(clock) {
	WS.Command("StartClock", [clock]);
}

function toTime(k, v) {
	return _timeConversions.msToMinSec(v);
}

function debugCallback(k, v) {
	var debug = $(".debug");
	debug.html(k + " = " + v + "<br />" + debug.html())
}
