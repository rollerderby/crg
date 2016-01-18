// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

$(init);

function init() {
	WS.Connect();
	WS.Register("Scoreboard.Snapshot(*)", snapshot);
}

function snapshot(k, v) {
	var idx = k.replace("Scoreboard.Snapshot(", "");
	idx = idx.substring(0, idx.indexOf(")"));
	var prefix = "Scoreboard.Snapshot(" + idx + ")";
	if (idx == 0)
		return;

	var period = WS.state[prefix + ".Clock(Period).Number"];
	var jam = WS.state[prefix + ".Clock(Jam).Number"];
	var state = WS.state[prefix + ".State"];
	row = findSnapshotRow(idx, period, jam, state, v != null);

	if (row && v == null) {
		// Remove row
		row.remove();
	} else if (row && v != null) {
		if (k == prefix + ".CanRevert")
			row.find(".CanRevert").text(isTrue(WS.state[prefix + ".CanRevert"]) ? "Yes" : "No");
		if (k == prefix + ".Length")
			row.find(".Length").text(timeComputerToHuman(WS.state[prefix + ".Length"]));
		var inProgress = isTrue(WS.state[prefix + ".InProgress"]);
		var lastInProgress = row.data("InProgress")
		if (startsWith(k, prefix + ".Clock(Period).") || inProgress != lastInProgress)
			stateClock(row, prefix, "Period", inProgress);
		if (startsWith(k, prefix + ".Clock(Jam).") || inProgress != lastInProgress)
			stateClock(row, prefix, "Jam", inProgress);
		if (startsWith(k, prefix + ".Clock(Lineup).") || inProgress != lastInProgress)
			stateClock(row, prefix, "Lineup", inProgress);
		if (startsWith(k, prefix + ".Clock(Timeout).") || inProgress != lastInProgress)
			stateClock(row, prefix, "Timeout", inProgress);
		if (startsWith(k, prefix + ".Clock(Intermission).") || inProgress != lastInProgress)
			stateClock(row, prefix, "Intermission", inProgress);

		row.data("InProgress", inProgress);
	}
}

function startsWith(a, b) {
	return a.substring(0, b.length) == b;
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

function findSnapshotRow(idx, period, jam, state, create) {
	while (idx.length < 5) {
		idx = '0' + idx;
	}
	var row = $(".Snapshot_" + idx)[0];
	if (row == null && create) {
		row = $("<tr>").addClass("Snapshot Snapshot_"+idx);
		row.append($("<td>").addClass("Index").text(Number(idx)));
		row.append($("<td>").addClass("PeriodJam").text(period + ' / ' + jam));
		row.append($("<td>").addClass("State").text(state));

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
			r = $(r);
			if (r.data('index') > row.data('index')) {
				r.after(row);
				inserted = true;
				return false;
			}
		});
		if (!inserted) {
			$(".StateHistory table tbody").append(row);
		}
	}

	return $(row);
}
