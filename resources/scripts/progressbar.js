function UpdateProgress(){
    done = document.body.scrollTop;
    max = (document.body.scrollHeight - document.body.offsetHeight);
    document.getElementById('progressbar').style.width = done / max * 100 + '%';
}

document.body.addEventListener('scroll', UpdateProgress)
