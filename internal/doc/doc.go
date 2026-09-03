// Package doc's YAML schema:
//
//	title: "My Document"        # rendered as the single H1
//	blocks:
//	  - type: heading            # heading | paragraph | list | code | quote | table | hr
//	    level: 2                 # heading only, 1-6
//	    text: "Section"          # heading | paragraph | quote
//	    ordered: false           # list only
//	    items: ["one", "two"]    # list only
//	    language: go             # code only
//	    headers: ["A", "B"]      # table only
//	    rows: [["1", "2"]]       # table only, each row must match len(headers)
package doc
