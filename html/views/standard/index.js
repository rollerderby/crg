$(initialize);

function initialize() {
	$("[sbCopyDiv]").each(function(idx, div) {
		div = $(div);
		div.html($(div.attr("sbCopyDiv")).html());
	});

	WS.Connect();
	WS.AutoRegister();

	// Set Styles
	WS.Register( SettingsBase+".SwapTeams", function (k, v) {
		$(".Team1").toggleClass("Left", !isTrue(v)).toggleClass("Right", isTrue(v));
		$(".Team2").toggleClass("Left", isTrue(v)).toggleClass("Right", !isTrue(v));
		$(".Team").toggleClass("Swapped", isTrue(v));
	});

	WS.Register( SettingsBase+".CurrentView", function(k, v) {
		if ($("div#"+v+".DisplayPane").length == 0) {
			v = "scoreboard";
		}
		$("div#video>video").each(function() { this.pause(); });
		$(".DisplayPane.Show").addClass("Hide");
		$(".DisplayPane").removeClass("Show");
		$("div#" + v + ".DisplayPane").addClass("Show");
		$("div#" + v + ".DisplayPane>video").each(function() { this.currentTime = 0; this.play(); });
	});

	WS.Register( SettingsBase+".Image", function(k, v) {
		$("div#image>img").attr("src", v);
	});
	WS.Register( SettingsBase+".Video", function(k, v) {
		$("div#video>video").attr("src", v);
	});
	WS.Register( SettingsBase+".CustomHtml", function(k, v) {
		$("div#html>iframe").attr("src", v);
	});

	WS.Register( [ SettingsBase+".BoxStyle",
		SettingsBase+".BackgroundStyle",
		SettingsBase+".HideJamTotals",
		SettingsBase+".SidePadding" ], function(k, v) {
			var boxStyle = WS.state[SettingsBase+".BoxStyle"];
			var backgroundStyle = WS.state[SettingsBase+".BackgroundStyle"];
			var showJamTotals = !isTrue(WS.state[SettingsBase+".HideJamTotals"]);
			var sidePadding = WS.state[SettingsBase+".SidePadding"];

			// change box_flat_bright to two seperate classes in order to reuse much of the css
			if (boxStyle == 'box_flat_bright')
				boxStyle = 'box_flat bright';

			$("body").removeClass();
			if (boxStyle != "" && boxStyle != null)
				$("body").addClass(boxStyle);
			if (backgroundStyle != "" && backgroundStyle != null)
				$("body").addClass(backgroundStyle);
			$("div#scoreboard").toggleClass("JamScore", showJamTotals);

			left = 0;
			right = 0;
			if (sidePadding != "" && sidePadding != null) {
				left = sidePadding;
				right = left;
			}
			$("div#scoreboard").css({ "left": left + "%", "width": (100 - left - right) + "%" });
			$(window).trigger("resize");

	});
}
