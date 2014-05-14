$(document).ready(function(){
	// from http://ppk.developpez.com/tutoriels/javascript/gestion-cookies-javascript/
	function createCookie(name,value,days) {
		if (days) {
			var date = new Date();
			date.setTime(date.getTime()+(days*24*60*60*1000));
			var expires = "; expires="+date.toGMTString();
		}
		else var expires = "";
		document.cookie = name+"="+value+expires+"; path=/";
	}

	function readCookie(name) {
		var nameEQ = name + "=";
		var ca = document.cookie.split(';');
		for(var i=0;i < ca.length;i++) {
			var c = ca[i];
			while (c.charAt(0)==' ') c = c.substring(1,c.length);
			if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
		}
		return null;
	}
	
	function eraseCookie(name) {
		createCookie(name,"",-1)
	}

	info = readCookie("auth-info")
	if (info != null) {
		$("#info").html('<button type="button" class="close" \
				data-dismiss="alert" aria-hidden="true">&times;</button>\
				'+info.replace(/_/g, " "))
		$("#info").addClass("alert alert-dismissable")
		if (info.indexOf("Error:") == 0) {
			$("#info").addClass("alert-danger")
		} else {
			$("#info").addClass("alert-info")
		}
		eraseCookie("auth-info")
	}
})