// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

$(init);

function init() {
	$(window).resize(function () {
		width = $(window).width();
		height = $(window).height();
	});
	width = $(window).width();
	height = $(window).height();
	tWidth = window.outerWidth;
	tHeight = window.outerHeight;
	offsetWidth = tWidth - width;
	offsetHeight = tHeight - height;
	window.resizeTo(1366 + offsetWidth, 768 + offsetHeight);


	WS.Connect();
	WS.AutoRegister();

	$("button").button();
	$(".buttonset").buttonset();

	$(["Period", "Jam", "Lineup", "Timeout", "Intermission"]).each(function(idx, clock) {
		WS.Register("Scoreboard.Clock("+clock+").Running", function(k, v) {
			$(".Clock."+clock).toggleClass("Running", isTrue(v));
		});
		WS.Register("Scoreboard.Clock("+clock+").Adjustable", function(k, v) {
			$(".Clock."+clock+" button").prop("disabled", !isTrue(v));
		});
	});

	WS.Register("Scoreboard.Snapshot(*)", snapshot);
	$(["1", "2"]).each(function(idx, t) {
		WS.Register("Scoreboard.Team("+t+").OfficialReviewRetained", function(k, v) {
			$(".Team"+t+" .OfficialReviewRetained").toggleClass("active", isTrue(v));
		});

		var teamEditor = createEditorDialog(t);
		$(".Team"+t+" .EditTeam button").click(function() {
			teamEditor.dialog("option", "title", WS.state["Scoreboard.Team("+t+").Name"] + " Editor");
			teamEditor.dialog("open");
		});
		WS.Register("Scoreboard.Team("+t+").Skater(*)", function(k, v) { skater(t, k, v); });

		$(["Jammer", "Pivot"]).each(function(idx, p) {
			var btn = $(".Team"+t+" ."+p+" Button.Box");
			var key = "Scoreboard.Team("+t+")."+p+".InBox";
			btn.click(function() { WS.Set(key, (!btn.hasClass("active")).toString()); });
			WS.Register(key, function(k, v) { btn.toggleClass("active", isTrue(v)); });
		});
	});
	WS.Register("Scoreboard.Jam(*)", jam);
	WS.Register("Scoreboard.State", function(k, v) {
		$(".MasterControls .Timeout").toggleClass("active", v == "OTO");
		$(".Team1 .Timeout").toggleClass("active", v == "TTO1");
		$(".Team1 .OfficialReview").toggleClass("active", v == "OR1");
		$(".Team2 .Timeout").toggleClass("active", v == "TTO2");
		$(".Team2 .OfficialReview").toggleClass("active", v == "OR2");
	});
}

function registerButtonCommand(select, command, data) {
	$(select).click(function() { console.log("COMMAND: ", command, data); WS.Command(command, data); });
}

function startClock(clock) {
	WS.Command("StartClock", [clock]);
}

function toTime(k, v) {
	return timeComputerToHuman(v, true);
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
			if (inserted)
				return;
			r = $(r);
			if (r.data('index') < row.data('index')) {
				r.before(row);
				inserted = true;
				return;
			}
		});
		if (!inserted) {
			$(".StateHistory table tbody").prepend(row);
		}
	}

	return $(row);
}

function skater(t, k, v) {
	var id = k.substring(k.indexOf("Skater(")+7);
	id = id.substring(0, id.indexOf(")"));

	var base = "Scoreboard.Team("+t+").Skater("+id+")";
	var field = k.substring(base.length+1);

	if (v == null) {
		if (field == "ID") {
			$(".Team"+t+ " select.Skater option").filterByData("key", id).remove();
			$(".Team"+t+ " tr.Skater").filterByData("key", id).remove();
		}
		return;
	}

	addSkater(t, id);
	var option = $(".Team"+t+ " select.Skater option").filterByData("key", id);
	var row = $(".Team"+t+ " tr.Skater").filterByData("key", id);

	if (field == "Number") {
		row.data("sort", v)
		row.find(".Number span").text(v);
		row.find(".Number input").val(v);

		option.text(v);
		option.data("sort", v);

		sort($(".Team"+t+ " select.Skater"), id);
		sort($(".Team"+t+ " table.Skaters tbody"), id);
	} else {
		if (field.substring(0, 7) != "BoxTrip") {
			row.find("."+field+" span").text(v);
			row.find("."+field+" input[type=text]").val(v);
			row.find("input."+field+"[type=checkbox]").prop("checked", isTrue(v));
		}
	}
}

function sort(p, id) {
	p.each(function(idx, p1) {
		p1 = $(p1);
		var elem = p1.children().filterByData("key", id);

		d1 = elem.data("sort");
		var inserted = false;
		p1.children().each(function(idx, child) {
			child = $(child);
			var d2 = child.data("sort");
			if (d2 != null && d1 < d2) {
				inserted = true;
				elem.insertBefore(child);
				return false;
			}
		});
		if (!inserted) {
			p1.append(elem);
		}
	});
}

function addSkater(t, id) {
	if ($(".Team"+t+" select.Skater option").filterByData("key", id).length == 0) {
		var o = $("<option>").data("key", id).val(id).appendTo($(".Team"+t+" select.Skater"));
	}
	if ($(".Team"+t+" .SkaterRows tr.Skater").filterByData("key", id).length == 0) {
		var tr = $(".Team"+t+".Editor table.Skaters tr.AddRow").clone();
		tr.removeClass("AddRow").data("key", id).addClass("Skater");

		tr.children().each(function(idx, elem) {
			elem = $(elem);
			var children = elem.children();
			children.detach();
			elem.append($("<span>"));
			elem.append($("<div>").append(children));
		});

		tr.click(function() {
			if (!tr.hasClass("Edit")) {
				tr.addClass("Edit");
				tr.find(".Number input").focus();
			}
		});

		var deleteButton = $("<button>").text("X").click(function() {
			WS.Command("Scoreboard.Team("+t+").DeleteSkater", id);
			return false;
		});
		tr.find(".Buttons").empty().append(deleteButton);
		tr.appendTo($(".Team"+t+" .SkaterRows"));
	}
}

function jam(k, v) {
	var ids = WS.ParseIDs(k);
	if (ids.length < 1)
		return;

	var id = ids[0];

	var base = "Scoreboard.Jam("+id+")";
	var field = k.substring(base.length+1);

	if (v == null) {
		if (field == "Jam") {
			$("table.ScoreKeeper tbody tr").filterByData("key", id).remove();
		}
		return;
	}

	addJam(base, id);
	var row = $("table.ScoreKeeper tbody tr").filterByData("key", id);

	if (field == "Jam" || field == "Period") {
		row.find("td.Jam span."+field).text(v);
	} else {
		row.find("."+field).text(v);
	}
}

function addJam(base, id) {
	if ($("table.ScoreKeeper tbody tr").filterByData("key", id).length == 0) {
		var tr = $(".Team1 table.ScoreKeeper tbody tr.Template").clone();
		tr.removeClass("Template").data("key", id).addClass("JamRow").data("key", id);
		tr.appendTo($("table.ScoreKeeper tbody"));
	}
}

function createEditorDialog(t) {
	var dialog = $(".Team"+t+" .Editor");
	dialog.addClass("Team"+t);

	var skaterAddRow = dialog.find("table.Skaters tr.AddRow");
	var addName = skaterAddRow.find("input.Name");
	var addNumber = skaterAddRow.find("input.Number");
	var addButton = $("button.Add");
	addName.add(addNumber).change(function(event) {
		addButton.button("option", "disabled", (!addName.val() || !addNumber.val()));
	});
	skaterAddRow.find("input").keyup(function(event) {
		if (!addButton.hasClass("disabled") && (13 == event.which)) // Enter
			addButton.click();
	});
	addButton.button("option", "disabled", true);

	skaterAddRow.find("button.Add").click(function() {
		var number = skaterAddRow.find("input.Number").val();
		var name = skaterAddRow.find("input.Name").val();
		var legalName = skaterAddRow.find("input.LegalName").val();
		var insuranceNumber = skaterAddRow.find("input.InsuranceNumber").val();
		var isAlt = skaterAddRow.find("input.IsAlt").prop("checked");
		var isCaptain = skaterAddRow.find("input.IsCaptain").prop("checked");
		var isAltCaptain = skaterAddRow.find("input.IsAltCaptain").prop("checked");
		var isBenchStaff = skaterAddRow.find("input.IsBenchStaff").prop("checked");

		var obj = {
			Name: name, Number: number, LegalName: legalName, InsuranceNumber: insuranceNumber,
			IsCaptain: isCaptain, IsAlt: isAlt, IsAltCaptain: isAltCaptain, IsBenchStaff: isBenchStaff
		};
		WS.NewObject("Scoreboard.Team("+t+").Skater", obj);

		skaterAddRow.find("input[type=text]").val("");
		skaterAddRow.find("input[type=checkbox]").prop("checked", false);
		addButton.button("option", "disabled", true);
		skaterAddRow.find(".Number").focus();
	});

	return dialog.dialog({
		title: "Team Editor",
		width: "1000px",
		modal: true,
		autoOpen: false,
		buttons: { Close: function() { $(this).dialog("close"); } },
	});
}
