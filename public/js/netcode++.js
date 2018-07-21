jQuery(document).ready(function(){
	let accordionsMenu = $('.cd-accordion-menu');

	if( accordionsMenu.length > 0 ) {
		
		accordionsMenu.each(function(){
			let accordion = $(this);
			//detect change in the input[type="checkbox"] value
			accordion.on('change', 'input[type="checkbox"]', function(){
				let checkbox = $(this);
				//console.log(checkbox.prop('checked'));
				( checkbox.prop('checked') ) ? checkbox.siblings('ul').attr('style', 'display:none;').slideDown(200) : checkbox.siblings('ul').attr('style', 'display:block;').slideUp(200);
			});
		});
	} 
  
  $( ".menu-button" ).click(function() {
    $('.sidebar').toggleClass('sidebar-close');
    $('.content').toggleClass('content_full-width');
    //$('.menu-icon').toggleClass('click-rotate');
  });
});



