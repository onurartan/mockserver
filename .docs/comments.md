### About some comments in the code
You can see comment keys with structures such as [DEV_LOG] [OPTIONAL_LOG], etc. This is a method developed by the mockserver developer to gain information about what needs to be done and what does not need to be done when reviewing the code for updates, etc., using functions such as these.

> # [IMP_FUNC] (IMPORTANT_FUNC) =>  Used for functions that perform important tasks within a project.

> ## [DEV_LOG] => This indicates the commented-out version of functions and log functions that we do not want normal users to see but that we will need information about during mock server development. This allows us to convert it to its normal state and use it when necessary.

> ### [OPTIONAL_LOG] => Indicates log functions that have been left in case they are needed.

> #### [USED_AI] [GENERATED_AI] => Indicates functions developed using artificial intelligence. These functions have been developed using 100% artificial intelligence. ()

> ##### // [Alternative=$desc] => is a special comment block used to document and store multiple possible solutions or behaviors within the code. The description is a short explanation of what this alternative does or in what situations it can be used.