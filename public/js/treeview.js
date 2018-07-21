$(".tree-view .checkbox input + label").on("click",function(){
 
  if($(this).parent().parent().hasClass("is-folder")) {

    console.log("is folder!");
    
    var checkFolder = $(this).closest(".is-folder");
    
    checkFolder.find(".checkbox").removeClass("partial-check");
    
    if($(this).parent().find("input[type='checkbox']").is(":checked")) {
      
      checkFolder.find("input[type='checkbox']").prop( "checked", false );
    
    } 
    else
    {
    
      checkFolder.find("input[type='checkbox']").prop( "checked", true ); 
    
    }
  
  }
  else
  {
    $(this).closest(".is-folder").find(">.checkbox").addClass("partial-check"); 
  }  
});

$(".tree-view").click(function(){
  
  var Folder = $(this).find(".is-folder");
  
  $.each(Folder, function(i, item){
    if($(this).find("> ul input[checked]")) {
      console.log("la carpeta " + i + " está llena");
    }
  });

});

$("a.folder-toggle").click(function(){
  var toggleTarget = $(this).closest(".is-folder")
  if(toggleTarget.hasClass("open")) {
      toggleTarget.removeClass("open");
  } else {
      toggleTarget.addClass("open");
  }
});