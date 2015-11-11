function timeComputerToHuman(t) {
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

	var ret = neg + min + ':' + pad(sec, 2) + '.' + subSec;
	console.log(ret, _t);
	return ret;
}

function timeHumanToComputer(h) {
}
