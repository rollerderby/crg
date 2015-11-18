// Sorry about the globals - couldn't get WS.whatever to be defined in setupPulsate()
var masterState = "";
var nT1TTO = 3;
var nT2TTO = 3;
var nT1OR = 1;
var nT2OR = 1;

$(initialize);

function initialize() {
	WS.Connect();
	WS.AutoRegister();

	var statusB = "";
	// Set Styles
	var view = "View";
	if (_windowFunctions.checkParam("preview", "true"))
		view = "Preview";
	WS.Register( "ScoreBoard.Setting(ScoreBoard." + view + "_SwapTeams)", function (k, v) {
		$(".Team1").toggleClass("Left", !isTrue(v)).toggleClass("Right", isTrue(v));
		$(".Team2").toggleClass("Left", isTrue(v)).toggleClass("Right", !isTrue(v));
		$(".Team").toggleClass("Swapped", isTrue(v));
	});

        WS.Register([
		"ScoreBoard.State",
		"ScoreBoard.Team(1).Timeouts",
		"ScoreBoard.Team(2).Timeouts",
		"ScoreBoard.Team(1).OfficialReviews",
		"ScoreBoard.Team(2).OfficialReviews"
		    ], function(k, v) {
			masterState = WS.state["ScoreBoard.State"];
	                if (masterState == "Jam")
	                      	statusB = "Jam";
  	            	else if (masterState == "Lineup")
  	                      statusB = "Lineup";
			else if (masterState == "TTO1" || masterState == "TTO2" || masterState == "OTO")
				statusB = "Timeout";
			else if (masterState == "OR1" || masterState == "OR2")
				statusB = "Review";
  	            	else if (masterState == "UnofficialFinal")
   	                     statusB = "Unofficial";
    	          	else if (masterState == "Final")
    	                    statusB = "Final";
    	          	else if (masterState == "Intermission")
    	                    statusB = "Halftime";
     	         	else if (masterState == "PreGame")
     	                   statusB = "Prebout";
                        $(".Status").removeClass("Show");
                        $(".Status.ShowIn" + statusB).addClass("Show");
			nT1TTO = WS.state["ScoreBoard.Team(1).Timeouts"];
			nT2TTO = WS.state["ScoreBoard.Team(2).Timeouts"];
			nT1OR = WS.state["ScoreBoard.Team(1).OfficialReviews"];
			nT2OR = WS.state["ScoreBoard.Team(2).OfficialReviews"];
			manageTeam1Images();
			manageTeam2Images();
        });

  	setupMainDiv($("#mainDiv"));
	// av.initialize();

    // Pulsate Timeouts if they're currently active. They'll be hidden in manageTimeoutImages
    $.each( [ 0, 1, 2 ], function(x, i) {
        setupPulsate(
                        function() { return (
                                        nT1TTO == i &&
					masterState == "TTO1");
				   },
                                $("#WftdaT1T"+(i+1)),
                                1000
                        );
        setupPulsate(
                        function() { return (
                                        nT2TTO == i &&
					masterState == "TTO2");
				   },
                       		$("#WftdaT2T"+(i+1)),
                        	1000
	                );
    });

    // Pulsate OR buttons.
    $.each( [ 1, 2 ], function(x, i) {
        setupPulsate(
                function() { return (
				masterState == "OR"+i);
			   },
                        $("#WftdaT"+i+"OR"),
                        1000
        );
    });

}

function setupMainDiv(div) {
  div.css({ position: "fixed" });

  $(window), "resize", function() {
    var aspect16x9 = _windowFunctions.get16x9Dimensions();
    div.css(aspect16x9).css("fontSize", aspect16x9.height);
  };
}

function manageTeam1Images() {
        // Called when something changes in relation to timeouts.
                // Have they one OR?
                if (nT1OR == 0) {
                        // Hide it
			  $("#WftdaT1OR").hide();
                } else {
                        // Show their OR Box
			  $("#WftdaT1OR").show();
                }
                // How's their timeouts looking?
                for ( var timeout = 1; timeout <= 3; timeout++ ) {
                        if (nT1TTO >= timeout )
				  $("#WftdaT1T"+timeout).show();
                        else
				  $("#WftdaT1T"+timeout).hide();
                }
}

function manageTeam2Images() {
        // Called when something changes in relation to timeouts.
                // Have they one OR?
                if (nT2OR == 0) {
                        // Hide it
			  $("#WftdaT2OR").hide();
                } else {
                        // Show their OR Box
			  $("#WftdaT2OR").show();
                }
                // How's their timeouts looking?
                for ( var timeout = 1; timeout <= 3; timeout++ ) {
                        if (nT2TTO >= timeout )
				  $("#WftdaT2T"+timeout).show();
                        else
				  $("#WftdaT2T"+timeout).hide();
                }
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
