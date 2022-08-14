/* global webroot */
/**
 * @typedef { import("./types/gochan").BoardThread } BoardThread
 * @typedef { import("./types/gochan").ThreadPost } ThreadPost
 */


import $ from "jquery";

import { getCookie } from "./cookies";
import { alertLightbox, promptLightbox } from "./lightbox";
import { getBooleanStorageVal, getNumberStorageVal } from "./storage";
import { isThreadWatched, unwatchThread, watchThread } from "./watcher";
import { openQR } from "./qr";

let doClickPreview = false;
let doHoverPreview = false;
let $hoverPreview = null;
let threadWatcherInterval = 0;

const threadRE = /^\d+/;
const videoTestRE = /\.(mp4)|(webm)$/;
const postrefRE = /\/([^\s/]+)\/res\/(\d+)\.html(#(\d+))?/;
const idRe = /^((reply)|(op))(\d+)/;
const opRegex = /\/res\/(\d+)(p(\d)+)?.html$/;


// data retrieved from /<board>/res/<thread>.json
/** @type {BoardThread} */
let currentThreadJSON = {
	posts: []
};

export function getPageThread() {
	let arr = opRegex.exec(window.location.pathname);
	let info = {
		board: currentBoard(),
		boardID: -1,
		op: -1,
		page: 0
	};
	if(arr === null) return info;
	if(arr.length > 1) info.op = arr[1];
	if(arr.length > 3) info.page = arr[3];
	if(arr.board != "") info.boardID = $("form#postform input[name=boardid]").val() -1;
	return info;
}

export function getUploadPostID(upload, container) {
	// if container, upload is div.upload-container
	// otherwise it's img or video
	let jqu = container? $(upload) : $(upload).parent();
	return insideOP(jqu) ? jqu.siblings().eq(4).text() : jqu.siblings().eq(3).text();
}

export function currentBoard() {
	let board = $("form#main-form input[type=hidden][name=board]").val();
	if(typeof board == "string")
		return board;
	return "";
}

export function currentThread() {
	// returns the board and thread ID if we are viewing a thread
	let thread = {board: currentBoard(), thread: 0};
	let splits = location.pathname.split("/");
	if(splits.length != 4)
		return thread;
	let reArr = threadRE.exec(splits[3]);
	if(reArr.length > 0)
		thread.thread = reArr[0];
	return thread;
}

/**
 * isPostVisible returns true if the post exists and is visible, otherwise false
 * @param {number} id the id of the post
 */
function isPostVisible(id) {
	let $post = $(`div#op${id}.op-post,div#reply${id}.reply`);
	if($post.length === 0)
		return false;
	return $post.find(".post-text").is(":visible");
}

/**
 * setPostVisibility sets the visibility of the post with the given ID. It returns true if it finds
 * a post or thread with the given ID, otherwise false
 * @param {number} id the id of the post to be toggled
 * @param {boolean} visibility the visibility to be set
 * @param onComplete called after the visibility is set
 */
function setPostVisibility(id, visibility, onComplete = () =>{}) {
	let $post = $(`div#op${id}.op-post, div#reply${id}.reply`);
	
	if($post.length === 0)
		return false;
	let $toSet = $post.find(".file-info,.post-text,.upload,.file-deleted-box,br");
	let $backlink = $post.find("a.backlink-click");
	if(visibility) {
		$toSet.show(0, onComplete);
		$post.find("select.post-actions option").each((e, elem) => {
			elem.text = elem.text.replace("Show", "Hide");
		});
		$backlink.text(id);
	} else {
		$toSet.hide(0, onComplete);
		$post.find("select.post-actions option").each((e, elem) => {
			elem.text = elem.text.replace("Hide", "Show");
		});
		$backlink.text(`${id} (hidden)`);
	}
	return true;
}

/**
 * setThreadVisibility sets the visibility of the thread with the given ID, as well as its replies.
 * It returns true if it finds a thread with the given ID, otherwise false
 * @param {number} id the id of the thread to be hidden
 * @param {boolean} visibility the visibility to be set
 */
function setThreadVisibility(opID, visibility) {
	let $thread = $(`div#op${opID}.op-post`).parent(".thread");
	if($thread.length === 0) return false;
	return setPostVisibility(opID, visibility, () => {
		let $toSet = $thread.find(".reply-container,b,br");
		if(visibility) {
			$toSet.show();
		} else {
			$toSet.hide();
		}
	});
}

export function insideOP(elem) {
	return $(elem).parents("div.op-post").length > 0;
}

/**
 * Formats the timestamp strings from JSON into a more readable format
 * @param {string} dateStr timestamp string, assumed to be in ISO Date-Time format
 */
function formatDateString(dateStr) {
	let date = new Date(dateStr);
	return date.toDateString() + ", " + date.toLocaleTimeString();
}

/**
 * Formats the given number of bytes into an easier to read filesize
 * @param {number} size
 */
function formatFileSize(size) {
	if(size < 1000) {
		return `${size} B`;
	} else if(size <= 100000) {
		return `${(size/1024).toFixed(1)} KB`;
	} else if(size <= 100000000) {
		return `${(size/1024/1024).toFixed(2)} MB`;
	}
	return `${(size/1024/1024/1024).toFixed(2)} GB`;
}

/**
 * creates an element from the given post data
 * @param {ThreadPost} post
 * @param {string} boardDir
 */
function createPostElement(post, boardDir, elementClass = "inlinepostprev") {
	let $post = $("<div/>")
		.prop({class: elementClass});
	$post.append(
		$("<a/>").prop({
			id: post.no.toString(),
			class: "anchor"
		}),
		$("<input/>")
			.prop({
				type: "checkbox",
				id: `check${post.no}`,
				name: `check${post.no}`
			}),
		$("<label/>")
			.prop({
				class: "post-info",
				for: `check${post.no}`
			}).append(formatDateString(post.time)),
		" ",
		$("<a/>")
			.prop({
				href: webroot + boardDir + "/res/" + ((post.resto > 0)?post.resto:post.no) + ".html#" + post.no
			}).text("No."),
		" ",
		$("<a/>")
			.prop({
				class: "backlink-click",
				href: `javascript:quote(${post.no})`
			}).text(post.no), "<br/>",
	);
	let $postInfo = $post.find("label.post-info");
	let postName = (post.name == "" && post.trip == "")?"Anonymous":post.name;
	let $postName = $("<span/>").prop({class: "postername"});
	if(post.email == "") {
		$postName.text(postName);
	} else {
		$postName.append($("<a/>").prop({
			href: "mailto:" + post.email
		}).text(post.name));
	}
	$postInfo.prepend($postName);
	if(post.trip != "") {
		$postInfo.prepend($postName, $("<span/>").prop({class: "tripcode"}).text("!" + post.trip), " ");
	} else {
		$postInfo.prepend($postName, " ");
	}

	if(post.sub != "")
		$postInfo.prepend($("<span/>").prop({class:"subject"}).text(post.sub), " ");


	if(post.filename != "" && post.filename != "deleted") {
		let thumbFile = getThumbFilename(post.tim);
		$post.append(
			$("<div/>").prop({class: "file-info"})
				.append(
					"File: ",
					$("<a/>").prop({
						href: webroot + boardDir + "/src/" + post.tim,
						target: "_blank"
					}).text(post.tim),
					` - (${formatFileSize(post.fsize)} , ${post.w}x${post.h}, `,
					$("<a/>").prop({
						class: "file-orig",
						href: webroot + boardDir + "/src/" + post.tim,
						download: post.filename,
					}).text(post.filename),
					")"
				),
			$("<a/>").prop({class: "upload-container", href: webroot + boardDir + "/src/" + post.tim})
				.append(
					$("<img/>")
						.prop({
							class: "upload",
							src: webroot + boardDir + "/thumb/" + thumbFile,
							alt: webroot + boardDir + "/src/" + post.tim,
							width: post.tn_w,
							height: post.tn_h
						})
				)	
		);
	}
	$post.append(
		$("<div/>").prop({
			class: "post-text"
		}).html(post.com)
	);
	addPostDropdown($post);
	return $post;
}

/**
 * Return the appropriate thumbnail filename for the given upload filename (replacing gif/webm with jpg, etc)
 * @param {string} filename
 */
function getThumbFilename(filename) {
	let nameParts = /([^.]+)\.([^.]+)$/.exec(filename);
	if(nameParts === null) return filename;
	let name = nameParts[1] + "t";
	let ext = nameParts[2];
	if(ext == "gif" || ext == "webm")
		ext = "jpg";

	return name + "." + ext;
}

export function updateThreadJSON() {
	let thread = currentThread();
	if(thread.thread === 0) return; // not in a thread
	return getThreadJSON(thread.thread, thread.board).then((json) => {
		if(!(json.posts instanceof Array) || json.posts.length === 0)
			return;
		currentThreadJSON = json;
	}).catch(e => {
		console.error(`Failed updating current thread: ${e}`);
		clearInterval(threadWatcherInterval);
	});
}

function updateThreadHTML() {
	let thread = currentThread();
	if(thread.thread === 0) return; // not in a thread
	let numAdded = 0;
	for(const post of currentThreadJSON.posts) {
		let selector = "";
		if(post.resto === 0)
			selector += `div#${post.no}.thread`;
		else
			selector += `a#${post.no}.anchor`;
		let elementExists = $(selector).length > 0;
		if(elementExists)
			continue; // TODO: check for edits
		
		let $replyContainer = $("<div/>").prop({
			id: `replycontainer${post.no}`,
			class: "reply-container"
		}).append(
			createPostElement(post, thread.board, "reply")
		);

		$replyContainer.appendTo(`div#${post.resto}.thread`);
		numAdded++;
	}
	if(numAdded === 0) return;
}

export function updateThread() {
	return updateThreadJSON().then(updateThreadHTML);
}

function createPostPreview(e, $post, inline = true) {
	let $preview = $post.clone();
	if(inline) $preview = addPostDropdown($post.clone());
	$preview
		.prop({class: "inlinepostprev"})
		.find("div.inlinepostprev").remove()
		.find("a.postref").on("click", expandPost);
	if(inline) {
		$preview.insertAfter(e.target);
	}
	initPostPreviews($preview);
	return $preview;
}

function previewMoveHandler(e) {
	if($hoverPreview === null) return;
	$hoverPreview.css({position: "absolute"}).offset({
		top: e.pageY + 8,
		left: e.pageX + 8
	});
}

function expandPost(e) {
	e.preventDefault();
	if($hoverPreview !== null) $hoverPreview.remove();
	let $next = $(e.target).next();
	if($next.prop("class") == "inlinepostprev" && e.type == "click") {
		// inline preview is already opened, close it
		$next.remove();
		return;
	}
	let href = e.target.href;
	let hrefArr = postrefRE.exec(href);
	if(hrefArr === null) return; // not actually a link to a post, abort
	let postID = hrefArr[4]?hrefArr[4]:hrefArr[2];

	let $post = $(`div#op${postID}, div#reply${postID}`).first();
	if($post.length > 0) {
		let $preview = createPostPreview(e, $post, e.type == "click");
		if(e.type == "mouseenter") {
			$hoverPreview = $preview.insertAfter(e.target);
			$(document.body).on("mousemove", previewMoveHandler);
		}
		return;
	}
	if(e.type == "click") {
		$.get(href, data => {
			$post = $(data).find(`div#op${postID}, div#reply${postID}`).first();
			if($post.length < 1) return; // post not on this page.
			createPostPreview(e, $post, true);
		}).catch((t, u, v) => {
			alertLightbox(v, "Error");
			return;
		});
	}
}

export function initPostPreviews($post = null) {
	if(getPageThread().board == "" && $post === null) return;
	doClickPreview = getBooleanStorageVal("enablepostclick", true);
	doHoverPreview = getBooleanStorageVal("enableposthover", false);
	let $refs = null;
	$refs = $post === null ? $("a.postref") : $post.find("a.postref");

	if(doClickPreview) {
		$refs.on("click", expandPost);
	} else {
		$refs.off("click", expandPost);
	}

	if(doHoverPreview) {
		$refs.on("mouseenter", expandPost).on("mouseleave", () => {
			if($hoverPreview !== null) $hoverPreview.remove();
			$hoverPreview = null;
			$(document.body).off("mousemove", previewMoveHandler);
		});
	} else {
		$refs.off("mouseenter").off("mouseleave").off("mousemove");
	}
}

export function prepareThumbnails() {
	// set thumbnails to expand when clicked
	$("a.upload-container").on("click", function(e) {
		e.preventDefault();
		let a = $(this);
		let thumb = a.find("img.upload");
		let thumbURL = thumb.attr("src");
		let uploadURL = thumb.attr("alt");
		thumb.removeAttr("width").removeAttr("height");

		var fileInfoElement = a.prevAll(".file-info:first");
		
		if(videoTestRE.test(thumbURL + uploadURL)) {
			// Upload is a video
			thumb.hide();
			var video = $("<video />")
			.prop({
				src: uploadURL,
				autoplay: true,
				controls: true,
				class: "upload",
				loop: true
			}).insertAfter(fileInfoElement);

			fileInfoElement.append($("<a />")
			.prop("href", "javascript:;")
			.on("click", function() {
				video.remove();
				thumb.show();
				this.remove();
				thumb.prop({
					src: thumbURL,
					alt: uploadURL
				});
			}).css({
				"padding-left": "8px"
			}).html("[Close]<br />"));
		} else {
			// upload is an image
			thumb.attr({
				src: uploadURL,
				alt: thumbURL
			});
		}
		return false;
	});
}

function selectedText() {
	if(!window.getSelection) return "";
	return window.getSelection().toString();
}

export function quote(no) {
	if(getBooleanStorageVal("useqr", true)) {
		openQR();
	}
	let msgboxID = "postmsg";	

	let msgbox = document.getElementById("qr" + msgboxID);
	if(msgbox === null)
		msgbox = document.getElementById(msgboxID);
	let selected = selectedText();
	let lines = selected.split("\n");

	if(selected !== "") {
		for(let l = 0; l < lines.length; l++) {
			lines[l] = ">" + lines[l];
		}
	}
	let cursor = (msgbox.selectionStart !== undefined)?msgbox.selectionStart:msgbox.value.length;
	let quoted = lines.join("\n");
	if(quoted != "") quoted += "\n";
	msgbox.value = msgbox.value.slice(0, cursor) + `>>${no}\n` +
		quoted + 
		msgbox.value.slice(cursor);
	
	if(msgbox.id == "postmsg")
		window.scroll(0,msgbox.offsetTop - 48);
	msgbox.focus();
}
window.quote = quote;

function handleActions(action, postIDStr) {
	let idArr = idRe.exec(postIDStr);
	if(!idArr) return;
	let postID = idArr[4];
	let board = currentBoard();
	switch(action) {
		case "Watch thread":
			watchThread(postID, board);
			break;
		case "Unwatch thread":
			unwatchThread(postID, board);
			break;
		case "Show thread":
			setThreadVisibility(postID, true);
			break;
		case "Hide thread":
			setThreadVisibility(postID, false);
			break;
		case "Show post":
			setPostVisibility(postID, true);
			break;
		case "Hide post":
			setPostVisibility(postID, false);
			break;
		case "Edit post":
			editPost(postID, board);
			break;
		case "Report post":
			reportPost(postID, board);
			break;
		case "Delete file":
			deletePost(postID, board, true);
			break;
		case "Delete thread":
		case "Delete post":
			deletePost(postID, board);
			break;
	}
}

export function addPostDropdown($post) {
	if($post.find("select.post-actions").length > 0)
		return $post;
	let $postInfo = $post.find("label.post-info");
	let isOP = $post.prop("class").split(" ").indexOf("op-post") > -1;
	let hasUpload = $postInfo.siblings("div.file-info").length > 0;
	let postID = $postInfo.parent().attr("id");
	let threadPost = isOP?"thread":"post";
	let $ddownMenu = $("<select />", {
		class: "post-actions",
		id: postID
	}).append("<option disabled selected>Actions</option>");
	let idNum = idRe.exec(postID)[4];
	if(isOP) {
		if(isThreadWatched(idNum, currentBoard())) {
			$ddownMenu.append("<option>Unwatch thread</option>");
		} else {
			$ddownMenu.append("<option>Watch thread</option>");
		}
	}
	let showHide = isPostVisible(idNum)?"Hide":"Show";
	$ddownMenu.append(
		`<option>${showHide} ${threadPost}</option>`,
		`<option>Edit post</option>`,
		`<option>Report post</option>`,
		`<option>Delete ${threadPost}</option>`,
	).insertAfter($postInfo)
	.on("change", e => {
		handleActions($ddownMenu.val(), postID);
		$ddownMenu.val("Actions");
	});
	if(hasUpload)
		$ddownMenu.append(`<option>Delete file</option>`);
	return $post;
}

export function editPost(id, board) {
	let cookiePass = getCookie("password");
	promptLightbox(cookiePass, true, () => {
		$("input[type=checkbox]").prop("checked", false);
		$(`input#check${id}`).prop("checked", true);
		$("input[name=edit_btn]").trigger("click");
	}, "Edit post");
}

export function reportPost(id, board) {
	promptLightbox("", false, ($lb, reason) => {
		if(reason == "" || reason === null) return;
		let xhrFields = {
			board: board,
			report_btn: "Report",
			reason: reason,
			json: "1"
		};
		xhrFields[`check${id}`] = "on";
		$.post(webroot + "util", xhrFields).fail(data => {
			let errStr = data.error;
			if(errStr == undefined)
				errStr = data.statusText;
			alertLightbox(`Report failed: ${errStr}`, "Error");
		}).done(() => {
			alertLightbox("Report sent", "Success");
		}, "json");
	}, "Report post");
}
window.reportPost = reportPost;

function deletePostFile(id) {
	let $elem = $(`div#op${id}.op-post, div#reply${id}.reply`);
	if($elem.length === 0) return;
	$elem.find(".file-info,.upload-container").remove();
	$("<div/>").prop({
		class: "file-deleted-box",
		style: "text-align: center;"
	}).text("File removed").insertBefore($elem.find("div.post-text"));
	alertLightbox("File deleted", "Success");
}

function deletePostElement(id) {
	let $elem = $(`div#op${id}.op-post`);
	if($elem.length > 0) {
		$elem.parent().next().remove(); // also removes the <hr> element after
		$elem.parent().remove();
	} else {
		$(`div#replycontainer${id}`).remove();
	}
}

export function deletePost(id, board, fileOnly) {
	let cookiePass = getCookie("password");
	promptLightbox(cookiePass, true, ($lb, password) => {
		let xhrFields = {
			board: board,
			boardid: $("input[name=boardid]").val(),
			delete_btn: "Delete",
			password: password,
			json: "1"
		};
		xhrFields[`check${id}`] = "on";
		if(fileOnly) {
			xhrFields.fileonly = "on";
		}
		$.post(webroot + "util", xhrFields).fail(data => {
			if(data !== "");
				alertLightbox(`Delete failed: ${data.error}`, "Error");
		}).done(data => {
			if(data.error == undefined || data == "") {
				if(location.href.indexOf(`/${board}/res/${id}.html`) > -1) {
					alertLightbox("Thread deleted", "Success");
				} else if(fileOnly) {
					deletePostFile(id);
				} else {
					deletePostElement(id);
				}
			} else {
				if(data.boardid === 0 && data.postid === 0) {
					alertLightbox(`Error deleting post #${id}: Post doesn't exist`, "Error");
				} else if(data !== "") {
					alertLightbox(`Error deleting post #${id}`, "Error");
					console.log(data);
				}
			}
		}, "json");
	}, "Password");
}
window.deletePost = deletePost;

export function getThreadJSON(threadID, board) {
	return $.ajax({
		url: `${webroot}${board}/res/${threadID}.json`,
		cache: false,
		dataType: "json"
	});
}

$(() => {
	let pageThread = getPageThread();
	if(pageThread.op < 1) return; // not in a thread

	threadWatcherInterval = setInterval(updateThread, getNumberStorageVal("watcherseconds", 10) * 1000);
});
