var showComposeBox = function () {
  $(".compose-box").show();
  $('.status-message').hide();
};

var closeComposeBox = function () {
  $(".compose-box").hide();
};

var tweet = function () {
  var text = $('.tweet-text').val();
  $.post('/ajaxupdate', {'text': text}).done(function () {
    closeComposeBox();
  }).fail(function (jqXHR) {
    $('.status-message').show();
    $('.status-message').text(jqXHR.responseText);
  });
};