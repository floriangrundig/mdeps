ELM_CORE_LIBS = "^(Dict)|^(String)|^(Time)|^(Task)|^(Regex)|^(Maybe)|^(Array)|^(Effects)|^(Signal)|^(Number)|^(List)|^(Date)|^(Exts)"
COLLAPSE_LIBS = "_Html_!!!^(Html)@@@_Json_!!!^(Json)@@@_Hop_!!!^(Hop)"

run:  
	@go run Main.go -n 3 \
    --ignoreDependencyRegEx=t$(ELM_CORE_LIBS) \
    --replaceDependencyRegEx=$(COLLAPSE_LIBS) \
    $(source)

basemodulepng: basemodule
	dot -Tsvg basemodules.dot -o basemodules.svg

basemodule:
	@go run Main.go -d -n 1 \
    --dotDiagramTitle="Base Modules" \
    --ignoreDependencyRegEx=$(ELM_CORE_LIBS) \
    --replaceDependencyRegEx=$(COLLAPSE_LIBS) \
    $(source) > basemodules.dot

basemodule-neo4j:
	@go run Main.go --neo4j -n 1 \
    --ignoreDependencyRegEx=$(ELM_CORE_LIBS) \
    --replaceDependencyRegEx=$(COLLAPSE_LIBS) \
    $(source) 
        
    
install-graphviz:
	brew install graphviz
