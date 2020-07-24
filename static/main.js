$(function(){
    var path = window.location.pathname;
    path = decodeURIComponent(path);
    initTable(path);
});
function getParentPath(path) {
    var m = path.split("/").slice(0);
    m.splice(m.length-1,1);
    var p = m.join("/");
    if(p == ""){
        p = "/";
    }
    if(path == "" || path == "/"){
        p = "";
    }
    return p;
}
function initTable(path) {
    $("#title").text("Index of "+path);
    $("#heading").text("Index of "+path);
    var p = getParentPath(path);
    if(p != ""){
        $("#table").append("<tr>"+
            "        <td class=\"file-name\">"+
            "            <a class=\"icon icon-up\" href=\""+p+"\">..</a>"+
            "        </td>"+
            "        <td class=\"file-size\"></td>"+
            "        <td class=\"file-date-modified\"></td>"+
            "       </tr>");
    }
    $.ajax({
        method: 'GET',
        url: '/data/data.json',
        success: function (data) {
            var html = "";
            var items = [];
            searchItem(path, data, items);
            $.each(items, function(i, item) {
                var n = "";
                var style = "icon-file";
                var a = item["downloadUrl"];
                if(item["isFolder"]){
                    style = "icon-dir";
                    n = "/";
                    if(path == "/"){
                        a = "/" + item["fileName"];
                    }else{
                        a = path + "/" + item["fileName"];
                    }
                }
                html+="<tr>"+
                    "    <td class=\"file-name\"><a class=\"icon "+style+"\" href=\""+a+"\">"+item["fileName"]+""+n+"</a></td>"+
                    "    <td class=\"file-size\">"+item["fileSize"]+"</td>"+
                    "    <td class=\"file-date-modified\">"+item["lastOpTime"]+"</td>"+
                    "</tr>";
            });
            $("#table").append(html);
        }
    });
}
function searchItem(path, data, items) {
    $.each(data, function(i, item) {
        if(path == item["path"]){
            items.push(item);
        }else{
             if(item["children"]!=null && item["children"].length > 0){
                 searchItem(path, item["children"], items);
             }
        }
    });
}