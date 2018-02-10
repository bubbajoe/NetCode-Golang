What to do next....

Priority: 100 = high priority, 0 = No priority/hope to do in the future

Database: MongoDB;

The folder structure is a list of nodes, Folders can have nodes. Files cannot have nodes
[{
    text: "Folder 1",
    nodes: [{
        text: "Folder 2",
        nodes: [
        {
            text: "File1"
        },
        {
            text: "File2.txt"
        }]
    },
    {
        text: "File2.txt"
    }]
},{
    text: "Folder 3"
},{
    text: "Folder 4"
}];

Folder Algorithm:
File Database Object
_id:1
filepath: folder 1/folder 2/
filename: file 1
folder_id: 2

Project
dir:[{text:"Folder 1",node:[{text:"Folder 2",node:[{text:"File 1"}]},{text:"File 2"}]},{text:"Folder 3"},{text:"Folder 4"}]
