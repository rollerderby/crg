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
	});
	width = $(window).width();
	height = $(window).height();
	tWidth = window.outerWidth;
	tHeight = window.outerHeight;
	offsetWidth = tWidth - width;
	offsetHeight = tHeight - height;
	window.resizeTo(1366 + offsetWidth, 768 + offsetHeight);
	

	$("[sbCopyDiv]").each(function(idx, div) {
		div = $(div);
		div.html($(div.attr("sbCopyDiv")).html());
	});
	$("button").button();
	$(".buttonset").buttonset();
	
	WS.Connect();
	WS.AutoRegister();

	$(["Period", "Jam", "Lineup", "Timeout", "Intermission"]).each(function(idx, clock) {
		WS.Register("ScoreBoard.Clock("+clock+").Running", function(k, v) {
			$(".Clock."+clock).toggleClass("Running", isTrue(v));
		});
		WS.Register("ScoreBoard.Clock("+clock+").Adjustable", function(k, v) {
			$(".Clock."+clock+" button").prop("disabled", !isTrue(v));
		});
	});

	WS.Register("ScoreBoard.Snapshot(*)", snapshot);
	$(["1", "2"]).each(function(idx, t) {
		WS.Register("ScoreBoard.Team("+t+").OfficialReviewRetained", function(k, v) {
			$(".Team"+t+" .OfficialReviewRetained").toggleClass("active", isTrue(v));
		});

		var teamEditor = createEditorDialog(t);
		$(".Team"+t+" .EditTeam button").click(function() {
			console.log("open team "+t+" editor")
			teamEditor.dialog("option", "title", WS.state["ScoreBoard.Team("+t+").Name"] + " Editor");
			teamEditor.dialog("open");
		});
		WS.Register("ScoreBoard.Team("+t+").Skater(*)", function(k, v) { skater(t, k, v); });
	});
	WS.Register("ScoreBoard.State", function(k, v) {
		$(".MasterControls .Timeout").toggleClass("active", v == "OTO");
		$(".Team1 .Timeout").toggleClass("active", v == "TTO1");
		$(".Team1 .OfficialReview").toggleClass("active", v == "OR1");
		$(".Team2 .Timeout").toggleClass("active", v == "TTO2");
		$(".Team2 .OfficialReview").toggleClass("active", v == "OR2");
	});

	$("#debugClocks").click(function() {
		WS.Command("Set", ["ScoreBoard.Clock(Period).Time.Max", "10000"]);
		WS.Command("Set", ["ScoreBoard.Clock(Jam).Time.Max", "5000"]);
		WS.Command("Set", ["ScoreBoard.Clock(Intermission).Time.Max", "45000"]);
		WS.Command("ScoreBoard.Reset");
	});
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
	var idx = k.replace("ScoreBoard.Snapshot(", "");
	idx = idx.substring(0, idx.indexOf(")"));
	var prefix = "ScoreBoard.Snapshot(" + idx + ")";
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

	var base = "ScoreBoard.Team("+t+").Skater("+id+")";
	if (k == base && v == null) {
		$(".Skater ."+id).remove();
		return;
	}

	addSkater(t, id);
	var select = $(".Team"+t+ " select.Skater option[value="+id+"]");
	var row = $(".Team"+t+ " tr.Skater[key="+id+"]");

	var field = k.substring(base.length+1);
	if (field == "Number") {
		row.find(".Number").text(v);
		select.text(v);
		select.prop("sort", v);
		sort($(".Team"+t+ " select.Skater"));
	} else {
		row.find("."+field).text(v);
	}
}

function sort(elem) {
}

function addSkater(t, id) {
	if ($(".Team"+t+" select.Skater option[value="+id+"]").length == 0) {
		$("<option>").val(id).appendTo($(".Team"+t+" select.Skater"));
	}
	if ($(".Team"+t+" .SkaterRows tr.Skater[key="+id+"]").length == 0) {
		var tr = $("<tr>").addClass("Skater").attr("key", id);
		$("<td>").addClass("Number").appendTo(tr);
		$("<td>").addClass("Name").appendTo(tr);
		$("<td>").addClass("InsuranceNumber").appendTo(tr);
		$("<td>").addClass("LegalName").appendTo(tr);
		$("<td>").addClass("Description").appendTo(tr);
		tr.appendTo($(".Team"+t+" .SkaterRows"));
	}
}

function createEditorDialog(t) {
	var dialog = $(".Team"+t+" .Editor");
	dialog.addClass("Team"+t);

	var skaterAddRow = dialog.find(".Skaters table tr.AddRow");
	skaterAddRow.find("button.Add").click(function() {
		var number = skaterAddRow.find(".Number").val();
		var name = skaterAddRow.find(".Name").val();
		var legalName = skaterAddRow.find(".LegalName").val();
		var insuranceNumber = skaterAddRow.find(".InsuranceNumber").val();
		var isAlt = skaterAddRow.find(".IsAlt").prop("checked");
		var isCaptain = skaterAddRow.find(".IsCaptain").prop("checked");
		var isAltCaptain = skaterAddRow.find(".IsAltCaptain").prop("checked");
		var isBenchStaff = skaterAddRow.find(".IsBenchStaff").prop("checked");

		var obj = {
			Name: name, Number: number, LegalName: legalName, InsuranceNumber: insuranceNumber, 
			IsCaptain: isCaptain, IsAlt: isAlt, IsAltCaptain: isAltCaptain, IsBenchStaff: isBenchStaff
		};
		console.log("Adding Skater", obj);
		WS.NewObject("ScoreBoard.Team("+t+").Skater", obj);
	});

	return dialog.dialog({
		title: "Team Editor",
		width: "1000px",
		modal: true,
		autoOpen: false,
		buttons: { Close: function() { $(this).dialog("close"); } },
	});
}
