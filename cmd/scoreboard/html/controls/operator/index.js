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

		WS.Register("ScoreBoard.Clock("+clock+").Running", function(k, v) {
			$(".Clock."+clock).toggleClass("Running", isTrue(v));
		});
		WS.Register("ScoreBoard.Clock("+clock+").Adjustable", function(k, v) {
			$(".Clock."+clock+" button").prop("disabled", !isTrue(v));
		});
	});

	registerButtonCommand(".MasterControls .StartJam", "StartJam");
	registerButtonCommand(".MasterControls .StopJam", "StopJam");
	registerButtonCommand(".MasterControls .Timeout", "Timeout");
	registerButtonCommand(".MasterControls .EndTimeout", "EndTimeout");
	registerButtonCommand(".MasterControls .Undo", "Undo");

	registerButtonCommand(".Team1 button.Timeout", "Timeout", ["TTO1"]);
	registerButtonCommand(".Team1 button.OfficialReview", "Timeout", ["OR1"]);
	registerButtonCommand(".Team2 button.Timeout", "Timeout", ["TTO2"]);
	registerButtonCommand(".Team2 button.OfficialReview", "Timeout", ["OR2"]);

	WS.Register("ScoreBoard.Snapshot(*)", snapshot);
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

function snapshot(k, v) {
	if (k.indexOf(".InProgress") == -1) {
		return
	}
	var idx = k.replace("ScoreBoard.Snapshot(", "");
	idx = idx.substring(0, idx.indexOf(")"));
	var prefix = "ScoreBoard.Snapshot(" + idx + ")";

	row = findSnapshotRow(idx, v != null);
	console.log("Looking for row", idx, row, v);

	if (row && v == null) {
		// Remove row
		console.log("REMOVING ROW");
		row.remove();
	} else if (row && v != null) {
		row.find(".Index").text(idx);
		row.find(".PeriodJam").text(
			WS.state[prefix + ".Clock(Period).Number"] + ' / ' + 
			WS.state[prefix + ".Clock(Jam).Number"]);
		row.find(".State").text(WS.state[prefix + ".State"]);
		row.find(".CanRevert").text(isTrue(WS.state[prefix + ".CanRevert"]) ? "Yes" : "No");
		row.find(".Length").text(timeComputerToHuman(WS.state[prefix + ".Length"]));
		var inProgress = isTrue(WS.state[prefix + ".InProgress"]);
		stateClock(row, prefix, "Period", inProgress);
		stateClock(row, prefix, "Jam", inProgress);
		stateClock(row, prefix, "Lineup", inProgress);
		stateClock(row, prefix, "Timeout", inProgress);
		stateClock(row, prefix, "Intermission", inProgress);
	}
}

function stateClock(row, prefix, clock, inProgress) {
	prefix = prefix + ".Clock(" + clock + ")";
	var td = row.find("." + clock);
	td.empty();
	if (isTrue(WS.state[prefix+".Running"]) || clock == 'Period') {
		td.append($("<span>").text("Start: " + timeComputerToHuman(WS.state[prefix+".StartTime"])));
		if (WS.state[prefix+".EndTime"] != null && !inProgress) {
			td.append($("<br />"));
			td.append($("<span>").text("End: " + timeComputerToHuman(WS.state[prefix+".EndTime"])));
		}
	}
}

function findSnapshotRow(idx, create) {
	while (idx.length < 5) {
		idx = '0' + idx;
	}
	var row = $(".Snapshot_" + idx)[0];
	if (row == null && create) {
		console.log("Snapshot " + idx + " not found, creating");
		row = $("<tr>").addClass("Snapshot Snapshot_"+idx);
		row.append($("<td>").addClass("Index"));
		row.append($("<td>").addClass("PeriodJam"));
		row.append($("<td>").addClass("State"));
		row.append($("<td>").addClass("CanRevert"));
		row.append($("<td>").addClass("StateClock Length"));
		row.append($("<td>").addClass("StateClock Period"));
		row.append($("<td>").addClass("StateClock Jam"));
		row.append($("<td>").addClass("StateClock Lineup"));
		row.append($("<td>").addClass("StateClock Timeout"));
		row.append($("<td>").addClass("StateClock Intermission"));
		row.data('index', idx);

		var inserted = false;
		$('.StateHistory table tbody tr').each(function (idx, r) {
			if (inserted)
				return;
			r = $(r);
			if (r.data('index') < row.data('index')) {
				console.log("<< Inserting " + row.data('index') + " before " + r.data('index'));
				r.before(row);
				inserted = true;
				return;
			}
		});
		if (!inserted) {
			console.log("Appending " + row.data('index') + " to end");
			$(".StateHistory table tbody").prepend(row);
		}
	}

	return $(row);
}
