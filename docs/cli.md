# CLI
This section describes the `scli` commands.

### List available commands

```bash
localhost:50051 > help
help <cmd> - prints help or a description by the command provided
list indexes <params> - allows to request index list
search <params> - run the search request across known indexes
```

### Show description of a command

```
localhost:50051 > help search
search <params> - returns the search results. It accepts the following params:

	text=<string> - the query text
	tags={"a":"a", "b":"b"} - the indexes with the tags values
	indexes=["index1", "index2"] - the list of indexes to run the search through
	distinct=<bool> - one record per index in the result
	limit=<int> - the number of records in the response
	as-table=<bool> - prints the result in a table form
```

### Execute a command

```bash
localhost:50051 > search text="lord OR honest" limit=10 as-table=true
Score  IdxId     Keywords       Segment                                                                
0....  test.txt  [lord]           EMILIA. [Within.] My lord, my lord! What, ho! my lord, my lord!      
0....  test.txt  [LORD lord]      LORD. Thou art a lord, and nothing but a lord.                       
0....  test.txt  [LORD lord]      FIRST LORD. It is the Count Rousillon, my good lord,                 
0....  test.txt  [LORD lord]      FIRST LORD. O my sweet lord, that you will stay behind us!           
0....  test.txt  [LORD lord]      SECOND LORD. Good my lord,                                           
0....  test.txt  [LORD lord]      SECOND LORD. Nay, good my lord, put him to't; let him have his way.  
0....  test.txt  [LORD lord]      SECOND LORD. On my life, my lord, a bubble.                          
0....  test.txt  [LORD lord]      SECOND LORD. Believe it, my lord, in mine own direct knowledge,      
0....  test.txt  [LORD lord]      FIRST LORD. You do not know him, my lord, as we do. Certain it is    
0....  test.txt  [LORD honest]    FIRST LORD. But you say she's honest.                                
Total:  3250
```
