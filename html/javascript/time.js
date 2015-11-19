// Copyright 2015 The CRG Authors (see AUTHORS file).
// All rights reserved.  Use of this source code is
// governed by a GPL-style license that can be found
// in the LICENSE file.

function timeComputerToHuman(t, s) {
	if (s == null)
		s = false;

	var _t = t;
	var pad = function(v) {
		v = v.toString();
		while (v.length < 2) {
			v = '0' + v;
		}
		return v;
	}

	t = Number(t);
	var neg = '';
	if (t < 0) {
		neg = '-';
		t = -t;
	}
	var ms = t % 1000;
	var subSec = Math.floor(ms / 100);
	t = (t - ms) / 1000;
	var sec = t % 60;
	t = (t - sec) / 60;
	var min = t;

	if (s)
		return ret = neg + min + ':' + pad(sec, 2);
	else
		return ret = neg + min + ':' + pad(sec, 2) + '.' + subSec;
}

function timeHumanToComputer(h) {
}
