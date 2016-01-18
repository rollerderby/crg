// Copyright 2015-2016 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

$(initialize);

function initialize() {
	WS.Connect();
	WS.AutoRegister();

	WS.Register([ "Scoreboard.State", "Scoreboard.Team(*).Timeouts", "Scoreboard.Team(*).OfficialReviews" ], function(k, v) {
		var statusB = "";
		var state = WS.state["Scoreboard.State"];
		if (state == "Jam")
			statusB = "Jam";
		else if (state == "Lineup")
			statusB = "Lineup";
		else if (state == "TTO1" || state == "TTO2" || state == "OTO")
			statusB = "Timeout";
		else if (state == "OR1" || state == "OR2")
			statusB = "Review";
		else if (state == "UnofficialFinal")
			statusB = "Unofficial";
		else if (state == "Final")
			statusB = "Final";
		else if (state == "Intermission")
			statusB = "Halftime";
		else if (state == "PreGame" || state == "")
			statusB = "Prebout";

		console.log('status', statusB);

		$(".Status").removeClass("Show");
		$(".Status.ShowIn" + statusB).addClass("Show");
	});

  	setupMainDiv($("#mainDiv"));
	// av.initialize();
}

function setupMainDiv(div) {
	div.css({ position: "fixed" });

	$(window), "resize", function() {
		var aspect16x9 = _windowFunctions.get16x9Dimensions();
		div.css(aspect16x9).css("fontSize", aspect16x9.height);
	};
}

var av = {
	videoElement: null,
	audioSource: '',
	videoSource: '',

	gotSources: function (sourceInfos) {
		for (var i = 0; i !== sourceInfos.length; ++i) {
			var sourceInfo = sourceInfos[i];
			console.log(sourceInfo);
			if (sourceInfo.kind === 'audio' && av.audioSource == "") {
				av.audioSource = sourceInfo.id;
			} else if (sourceInfo.kind === 'video' && ( i == 2 ) ) {
				av.videoSource = sourceInfo.id;
				console.log('Selecting 2nd camera');
			} else {
				console.log('Some other kind of source: ', sourceInfo);
			}
		}
		av.start();
	},

	initialize: function() {
		if (typeof MediaStreamTrack === 'undefined'){
			alert('This browser does not support MediaStreamTrack.\n\nTry Chrome Canary.');
		} else {
			MediaStreamTrack.getSources(av.gotSources);
			// av.start();
		}
	},

	successCallback: function(stream) {
		window.stream = stream; // make stream available to console

		if (av.videoElement == null) {
			av.videoElement = document.createElement("video");
			av.videoElement.className = 'video_underlay';
			document.body.appendChild(av.videoElement);
		}

		av.videoElement.src = window.URL.createObjectURL(stream);
		av.videoElement.play();
		$(document.body).addClass("HasUnderlay");
	},

	errorCallback: function(error) {
		console.log('navigator.getUserMedia error: ', error);
	},

	start: function() {
		if (!!window.stream) {
			videoElement.src = null;
			window.stream.stop();
		}
		var constraints = {
			audio: {
				optional: [{sourceId: av.audioSource}]
			},
			video: {
				optional: [{sourceId: av.videoSource}]
			},
			width: {min: 640, ideal: window.innerWidth},
			height: {min: 480, ideal: window.innerHeight},
			aspectRatio: 1.5,
		};
		console.log(constraints);
		console.log(window);

		var getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia;
		navigator.webkitGetUserMedia(constraints, av.successCallback, av.errorCallback);

	}
}
