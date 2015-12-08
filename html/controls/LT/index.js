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
	$(".Tabs").tabs();

	WS.Register("Scoreboard.Team(*).Skater(*)", skater);
	WS.Register("Scoreboard.Jam(*)", jam);
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

function skater(k, v) {
	var ids = WS.ParseIDs(k);
	if (ids.length < 2)
		return;

	var t = ids[0];
	var id = ids[1];

	var base = "Scoreboard.Team("+t+").Skater("+id+")";
	var field = k.substring(base.length+1);

	if (v == null) {
		if (field == "ID") {
			$(".Team"+t+ " tr.Skater").filterByData("key", id).remove();
		}
		return;
	}

	addSkater(base, t, id);
	var row = $(".Team"+t+ " tr.Skater").filterByData("key", id);

	if (field == "Number") {
		row.data("sort", v)
		row.find(".Number").text(v);

		sort($(".Team"+t+ " table.Skaters tbody"), id);
	} else if (field == "Position") {
		row.find("button.Position").removeClass("Active");
		row.find("button.Position."+v).addClass("Active");
	} else if (field == "InBox") {
		row.find("button.Box").toggleClass("Active", isTrue(v));
	} else {
		row.find("."+field).text(v);
	}
}

function addSkater(base, t, id) {
	if ($(".Team"+t+" .Skaters tbody tr.Skater").filterByData("key", id).length == 0) {
		var tr = $(".Team"+t+" table.Skaters tr.Template").clone();
		tr.removeClass("Template").data("key", id).addClass("Skater");
		tr.appendTo($(".Team"+t+" table.Skaters tbody"));

		tr.find("button.Bench").click(function() { WS.Set(base+".Position", "Bench"); });
		tr.find("button.Jammer").click(function() { WS.Set(base+".Position", "Jammer"); });
		tr.find("button.Pivot").click(function() { WS.Set(base+".Position", "Pivot"); });
		tr.find("button.Blocker").click(function() { WS.Set(base+".Position", "Blocker"); });
		tr.find("button.Box").click(function() { WS.Set(base+".InBox", (!tr.find("button.Box").hasClass("Active")).toString()); });
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
			$("table.PaperWork tbody tr").filterByData("key", id).remove();
		}
		return;
	}

	addJam(base, id);
	var row = $("table.PaperWork tbody tr").filterByData("key", id);

	if (field == "Jam" || field == "Period") {
		row.find("td.Jam span."+field).text(v);
	} else {
		row.find("."+field).text(v);
	}
}

function addJam(base, id) {
	if ($("table.PaperWork tbody tr").filterByData("key", id).length == 0) {
		var tr = $(".Team1 table.PaperWork tbody tr").clone();
		tr.removeClass("Template").data("key", id).addClass("JamRow").attr("boo", id);
		tr.appendTo($("table.PaperWork tbody"));
		console.log("Adding Jam("+id+")");
		// sort($(".Team1 table.PaperWork tbody tr"), id);
		// sort($(".Team2 table.PaperWork tbody tr"), id);
	}
}
