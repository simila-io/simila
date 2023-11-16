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

```bash
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
localhost:50051 > search text="lord sir honest" limit=10 as-table=true
Score  IdxId     Keywords       Segment                                                               
0.75   test.txt  [honest lord]    KATHARINE. I thank you, honest lord. Remember me                    
0.75   test.txt  [Honest lord]    PROSPERO.  [Aside]  Honest lord,                                    
0.75   test.txt  [honest lord]      Poor honest lord, brought low by his own heart,                   
0.63   test.txt  [Honest lord]    Pol. Honest, my lord?                                               
0.63   test.txt  [Honest lord]    IAGO. Honest, my lord?                                              
0.62   test.txt  [LORD Sir]       SECOND LORD. Sir, his wife, some two months since, fled from his    
0.62   test.txt  [LORD Sir]       FIRST LORD. Sir, I would advise you to shift a shirt; the violence  
0.62   test.txt  [LORD Sir]       FIRST LORD. Sir, as I told you always, her beauty and her brain go  
0.57   test.txt  [lord honest]      How far hence is thy lord, mine honest fellow?                    
0.56   test.txt  [Lord]           CLOWN. O Lord, sir!-There's a simple putting off. More, more, a     
Total:  3656
```
