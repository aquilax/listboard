// https://hacks.mozilla.org/2011/03/the-shortest-image-uploader-ever/
function upload(file) {
	/* Is the file an image? */
	if (!file || !file.type.match(/image.*/)) return;

	/* Lets build a FormData object*/
	var fd = new FormData(); // I wrote about it: https://hacks.mozilla.org/2011/01/how-to-develop-a-html5-image-uploader/
	fd.append("image", file); // Append the file
	var xhr = new XMLHttpRequest(); // Create the XHR (Cross-Domain XHR FTW!!!) Thank you sooooo much imgur.com
	xhr.open("POST", "https://api.imgur.com/3/upload.json"); // Boooom!
	xhr.setRequestHeader('Authorization', 'Client-ID c2e15b62bf762a8');
	xhr.onload = function() {
			// Big win!
			var data = JSON.parse(xhr.responseText),
				textarea = document.getElementById('textarea'),
				texta = textarea.value.split("\n");
			if (data.data && data.data.link) {
				link = "![Image](" + data.data.link + ")";
				texta.push(link);
				textarea.value = texta.join("\n");
			}
			//document.querySelector("#link").href = JSON.parse(xhr.responseText).upload.links.imgur_page;
		}
		/* And now, we send the formdata */
	xhr.send(fd);
}

function quote() {
	var textComponent = document.getElementById('textarea'),
		text = textComponent.value,
		startPos = textComponent.selectionStart,
		endPos = textComponent.selectionEnd,
		selLenght,
		segment,
		newSegment = [];
	if (startPos != undefined) {
		selLength = endPos - startPos;
		segment = text.substr(startPos, selLength);
		segment.split("\n").forEach(function(row) {
			newSegment.push("> " + row);
		});
		textComponent.value = [
			text.substr(0, startPos),
			newSegment.join("\n"),
			text.substr(endPos)
		].join("\n");
	}
}

// Find all the YouTube video embedded on a page
var videos = document.getElementsByClassName("youtube");

for (var i = 0; i < videos.length; i++) {

	var youtube = videos[i];
	var videoId = youtube.getAttribute('data-video');

	// Attach an onclick event to the YouTube Thumbnail
	youtube.onclick = function() {

		// Create an iFrame with autoplay set to true
		var iframe = document.createElement("iframe");
		iframe.setAttribute("src",
			"https://www.youtube.com/embed/" + videoId + "?autoplay=1&autohide=1&border=0&wmode=opaque&enablejsapi=1");

		// The height and width of the iFrame should be the same as parent
		iframe.className = 'video_frame'
			// Replace the YouTube thumbnail with YouTube HTML5 Player
		this.parentNode.replaceChild(iframe, this);

	};
}

function rsz(elem, max) {
	if (elem == undefined || elem == null) return false;
	if (max == undefined) max = 320;
	if (elem.width > elem.height) {
		if (elem.width > max) elem.width = max;
	} else {
		if (elem.height > max) elem.height = max;
	}
}

function tgl(elem, max) {
	if (elem == undefined || elem == null) return false;
	if (elem.alt != 1) {
		elem.removeAttribute('width');
		elem.removeAttribute('height');
		elem.alt = 1;
	} else {
		elem.removeAttribute('alt');
		rsz(elem, max);
	}
}
