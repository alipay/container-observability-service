package model

const (
	TPL = `
<!DOCTYPE html>
<html lang="en">
 <head> 
  <meta charset="UTF-8" /> 
  <meta name="viewport" content="width=device-width, initial-scale=1.0" /> 
  <meta http-equiv="X-UA-Compatible" content="ie=edge" /> 
  <title>Document</title> 

   <link href="/statics/reset.css" rel="stylesheet" /> 
   <link href="/statics/app.css" rel="stylesheet" /> 
   <link href="/statics/jsonTree.css" rel="stylesheet" /> 
   <link href="/statics/css_family.css" rel="stylesheet" /> 

 </head> 
 <body> 
  <header id="header"> 
   <nav id="nav" class="clearfix"> 
    <ul class="menu menu_level1"> 
     <li data-action="expand" class="menu__item" id="expand-all"> <span class="menu__item-name" style="text-decoration: underline;">展开内容>>></span> </li>
     <li data-action="collapse" class="menu__item" id="collapse-all"> <span class="menu__item-name" style="text-decoration: underline;">折叠内容>>></span> </li>
    </ul>
   </nav> 
   <div id="coords"></div> 
   <div id="debug"></div> 
  </header> 
  <div id="wrapper"> 
   <div id="tree"></div> 
  </div> 
  
  <script src="/statics/jsonTree.js"></script> 

  <script>
	// Get DOM-element for inserting json-tree
	var wrapper = document.getElementById("wrapper");
	var data = %s
	var tree = jsonTree.create(data, wrapper);
	document.getElementById("expand-all").addEventListener("click", function() {
  		tree.expand();
	})
	document.getElementById("collapse-all").addEventListener("click", function() {
  		tree.collapse();
	})

</script>  
 </body>
</html>
`
)
