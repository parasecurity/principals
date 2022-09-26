const { execSync } = require("child_process");
/*
var mirrname = execSync("ovs-vsctl show | grep dgadtc | awk 'NR==1 {print $2}'", (error) => {
    if (error) {
        console.log(`error: ${error.message}`);
        return;
    }
});
*/
let a =`ovs-vsctl \
-- --id=@p get port  ${mirrname} \
-- --id=@m create mirror name=m0 select-all=true output-port=@p \
-- set bridge br-int mirrors=@m`

a=a.replace(/\r?\n|\r/g, " ");

/*
execSync(a, (error) => {
    if (error) {
        console.log(`error: ${error.message}`);
        return;
    }
});*/

console.log("hey")


