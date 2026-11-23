document.querySelectorAll('a[href^="#"]').forEach(function(a){
  a.addEventListener('click',function(e){
    var id=this.getAttribute('href');
    var target=document.querySelector(id);
    if(target){e.preventDefault();target.scrollIntoView({behavior:'smooth'});}
  });
});

var sections=document.querySelectorAll('section[id]');
var links=document.querySelectorAll('.menu a');
window.addEventListener('scroll',function(){
  var pos=window.scrollY+120;
  sections.forEach(function(s){
    if(pos>=s.offsetTop && pos<s.offsetTop+s.offsetHeight){
      links.forEach(function(l){
        l.style.color=l.getAttribute('href')==='#'+s.id?'#00b4d8':'#ddd';
      });
    }
  });
});
