// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

var view = "View";
if (_windowFunctions.checkParam("preview", "true"))
	view = "Preview";
var SettingsBase = "Settings." + view;
$(function() {
	WS.Register(SettingsBase, function() {});
	WS.Register(
		["Scoreboard.State", "Scoreboard.Team(*).Timeouts", "Scoreboard.Team(*).OfficialReviews", "Scoreboard.Team(*).OfficialReviewRetained"],
		timeoutDisplay);
});

function jammer(k, v) {
	id = getTeamId(k);
	var prefix = "Scoreboard.Team(" + id + ").";
	var jammerName = WS.state[prefix + "Jammer.Name"];
	var lead = (WS.state[prefix + "Lead"] === "Lead");

	if (jammerName == null)
		jammerName = lead ? "Lead" : "";

	$(".Team" + id + " .Lead").toggleClass("HasLead", lead);
	$(".Team" + id).toggleClass("HasJammerName", (jammerName != "" && jammerName != null));
	return jammerName;
}

function getTeamId(k) {
	if (k.indexOf("Team(1)") > 0)
		return "1";
	if (k.indexOf("Team(2)") > 0)
		return "2";
	return null;
}

function nameUpdate(k, v) {
	id = getTeamId(k);
	var prefix = "Scoreboard.Team(" + id + ").";
	var name = WS.state[prefix + "Name"];
	var altName1 = WS.state[prefix + "AlternateName(overlay)"];
	var altName2 = WS.state[prefix + "AlternateName(scoreboard)"];

	if (altName1 != null && altName1 != "")
		name = altName1;
	else if (altName2 != null && altName2 != "")
		name = altName2;

	$(".Team" + id).toggleClass("HasName", name != "");
	return name;
}

function logoUpdate(k, v) {
	id = getTeamId(k);
	var prefix = "Scoreboard.Team(" + id + ").";
	var logo = WS.state[prefix + "Logo"];
	if (logo == null)
		logo = "";
	if (logo != "")
		logo = 'url("' + logo + '")';

	$(".Team" + id + ">.Logo").css("background-image", logo);
	$(".Team" + id).toggleClass("HasLogo", logo != "");
	var nameAutoFit = $(".Team" + id + ">.Name>div").data("AutoFit");
	if (nameAutoFit)
		nameAutoFit();
}

function timeoutDisplay(k, v) {
	var state = WS.state["Scoreboard.State"];

	for (var id = 1; id <= 2; id++) {
		var tto = WS.state["Scoreboard.Team(" + id + ").Timeouts"];
		var tor = WS.state["Scoreboard.Team(" + id + ").OfficialReviews"];
		var tror = isTrue(WS.state["Scoreboard.Team(" + id + ").OfficialReviewRetained"]);
		$(".Team" + id + " .Timeout1").toggleClass("Used", tto < 1);
		$(".Team" + id + " .Timeout2").toggleClass("Used", tto < 2);
		$(".Team" + id + " .Timeout3").toggleClass("Used", tto < 3);
		$(".Team" + id + " .OfficialReview1").toggleClass("Used", tor < 1);
		$(".Team" + id + " .OfficialReview1").toggleClass("Retained", tror);
	}

	$(".Team .Dot").removeClass("Active");

	if (state == "TTO1" || state == "TTO2") {
		var t = state.substring(3, 4);
		var dotSel = ".Team" + t + " .Timeout" + (Number(WS.state["Scoreboard.Team(" + t + ").Timeouts"]) + 1);
		$(dotSel).addClass("Active");
	}
	if (state == "OR1" || state == "OR2") {
		var t = state.substring(2, 3);
		var dotSel = ".Team" + t + " .OfficialReview1";
		$(dotSel).addClass("Active");
	}
}

function smallDescriptionUpdate(k, v) {
	var state = WS.state["Scoreboard.State"];

	$(".Clock.Description,.Team>.Timeouts,.Team>.OfficialReviews").removeClass("Red");
	if (state == "Lineup")
		return WS.state["Scoreboard.Clock(Lineup).Name"];
	if (state == "OTO") {
		$(".Clock.Description").addClass("Red");
		return WS.state["Scoreboard.Clock(Timeout).Name"];
	}

	if (state == "TTO1" || state == "TTO2") {
		var t = state.substring(3, 4);
		$(".Team" + t + ">.Timeouts").addClass("Red");
		$(".Clock.Description").addClass("Red");
		return "Team Timeout";
	}
	if (state == "OR1" || state == "OR2") {
		var t = state.substring(2, 3);
		$(".Team" + t + ">.OfficialReviews:not(.Header)").addClass("Red");
		$(".Clock.Description").addClass("Red");
		return "Official Review";
	}

	return '';
}

function intermissionDisplay() {
	var state = WS.state["Scoreboard.State"];

	ret = WS.state[SettingsBase+".Intermission(" + state + ")"];
	$(".Clock.Intermission .Time").toggleClass("Hide", state == "UnofficialFinal" || state == "Final");
	return ret;
}

function toClockInitialNumber(k, v) {
	var ret = '';
	$.each(["Period", "Jam"], function (i, c) {
		if (k.indexOf("Clock(" + c + ")") > -1) {
			var name = WS.state["Scoreboard.Clock(" + c + ").Name"];
			var number = WS.state["Scoreboard.Clock(" + c + ").Number.Num"];

			if (name != null && number != null)
				ret = name.substring(0, 1) + number;

			if (name == 'Period' && WS.state['Scoreboard.Clock(Period).Number.Max'] == 1)
				ret = 'Game';
		}
	});
	return ret;
}

function toTime(k, v) {
	return timeComputerToHuman(v, true);
}

function toInitial(k, v) {
	return v == null ? '' : v.substring(0, 1);
}

function clockRunner(k,v) {
	var lc = WS.state["Scoreboard.Clock(Lineup).Running"];
	var tc = WS.state["Scoreboard.Clock(Timeout).Running"];
	var ic = WS.state["Scoreboard.Clock(Intermission).Running"];

	var clock = "Jam";
	if (isTrue(tc))
		clock = "Timeout";
	else if (isTrue(lc))
		clock = "Lineup";
	else if (isTrue(ic))
		clock = "Intermission";

	$(".Clock,.SlideDown").removeClass("Show");
	$(".SlideDown.ShowIn" + clock + ",.Clock.ShowIn" + clock).addClass("Show");
}


// Show Clocks
WS.Register( "Scoreboard.Clock(*).Running", clockRunner );
WS.Register( 'Scoreboard.Clock(Period).Number.Max' );
