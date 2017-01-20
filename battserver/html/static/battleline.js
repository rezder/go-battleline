var batt={};
(function(batt){

    function createSvg(document){
        const COLORS_Troops=["#00ff00","#ff0000","#af3dff","#ffff00","#007fff","#ffa500"];
        const COLORS_Names=["Green","Red","Purpel","Yellow","Blue","Orange"];
        const COLOR_CardFrame="#ffffff";
        const COLOR_CardFrameSpecial="#000000";
        const COLOR_CardFrameLastMove="#0000cc";

        const ID_Cone=1;
        const ID_FlagTroop=2;
        const ID_FlagTac=3;
        const ID_Card=4;
        const ID_DeckTac=5;
        const ID_DeckTroop=6;
        const ID_DishTac=7;
        const ID_DishTroop=8;
        const ID_Hand=9;

        const TROOP_NO = 60;
	      const TAC_NO   = 10;

	      const TC_Alexander = 70;
	      const TC_Darius    = 69;
	      const TC_8         = 68;
	      const TC_123       = 67;
	      const TC_Fog       = 66;
	      const TC_Mud       = 65;
	      const TC_Scout     = 64;
	      const TC_Redeploy  = 63;
	      const TC_Deserter  = 62;
	      const TC_Traitor   = 61;

        let exp={};
        let card={};
        exp.card={};

	      exp.card.TC_Alexander = 70;
	      exp.card.TC_Darius    = 69;
	      exp.card.TC_8         = 68;
	      exp.card.TC_123       = 67;
	      exp.card.TC_Fog       = 66;
	      exp.card.TC_Mud       = 65;
	      exp.card.TC_Scout     = 64;
	      exp.card.TC_Redeploy  = 63;
	      exp.card.TC_Deserter  = 62;
	      exp.card.TC_Traitor   = 61;

        let hand={};
        exp.hand={};
        hand.vSpace=26;
        hand.hSpace=20;
        let flag={};
        exp.flag={};
        let cone={};
        exp.cone={};
        let click={};
        let id={};
        exp.id={};
        exp.id.ID_Cone=1;
        exp.id.ID_FlagTroop=2;
        exp.id.ID_FlagTac=3;
        exp.id.ID_Card=4;
        exp.id.ID_DeckTac=5;
        exp.id.ID_DeckTroop=6;
        exp.id.ID_DishTac=7;
        exp.id.ID_DishTroop=8;
        exp.id.ID_Hand=9;

        id.typeToName=function(type,no,player){
            let pText;
            if(player){
                pText="p";
            }else{
                pText="o";
            }
            let idName;
            switch (type){
            case ID_Cone:
            idName="k"+no+"Path";
                break;
            case ID_FlagTroop:
                idName=pText+"F"+no+"TroopGroup";
                break;
            case ID_FlagTac:
                idName=pText+"F"+no+"TacGroup";
                break;
            case ID_Card:
                idName="card"+no;
                break;
            case ID_DeckTac:
                idName="deckTacGroup";
                break;
            case ID_DeckTroop:
                idName="deckTroopGroup";
                break;
            case ID_DishTac:
                idName=pText+"DishTacGroup";
                break;
            case ID_DishTroop:
                idName=pText+"DishTroopGroup";
                break;
            case ID_Hand:
                idName=pText+"Hand";
                break;
            }
            return idName;
        };
        id.fromName=function from(idName){
            let res={};
            switch (idName.charAt(0)){
            case "k":
                res.type=ID_Cone;
                res.no=parseInt(idName.match(/\d/)[0]);
                break;
            case "p":
                let dish=false;
                let hand=false;
                if (idName.search(/Dish/)>-1){
                    dish=true;
                }else if (idName.search(/Hand/)>-1){
                    hand=true;
                }
                if (hand){
                    res.type=ID_Hand;
                }else{
                    if (idName.search(/Troop/)===-1){
                        if (dish){
                            res.type=ID_DishTac;
                        }else{
                            res.type=ID_FlagTac;
                        }
                    }else{
                        if (dish){
                            res.type=ID_DishTroop;
                        }else{
                            res.type=ID_FlagTroop;
                        }
                    }
                    if(!dish){
                        res.no=parseInt(idName.match(/\d/)[0]);
                    }
                }
                res.player=true;
                break;
            case "o":
                if (idName.search(/Troop/)===-1){
                    res.type=ID_FlagTac;
                }else{
                    res.type=ID_FlagTroop;
                }
                res.no=parseInt(idName.match(/\d/)[0]);
                res.player=false;
                break;
            case "c":
                res.type=ID_Card;
                res.no=parseInt(idName.match(/\d{1,2}/)[0]);
                break;
            case "d":
                if (idName.search(/Troop/)===-1){
                    res.type=ID_DeckTac;
                }else{
                    res.type=ID_DeckTroop;
                }
                break;
            }
            return res;
        };
        exp.id.fromName=id.fromName;

        class Area{
            constructor(x0,x1,y0,y1){
                this.x0=x0;
                this.x1=x1;
                this.y0=y0;
                this.y1=y1;
            }
            hit(x,y){
                let res=false;
                if(this.x0 <= x && this.x1 >=x && this.y0 <= y && this.y1 >= y){
                    res=true;
                }
                return res;
            }
        }
        card.isTac=function(cardix){
            return cardix>TROOP_NO;
        };
        exp.card.isTac=card.isTac;
        card.count=function(group){
            let res={n:0,x:0,y:0};
            for (let i=0;i<group.childNodes.length;i++){
                let node=group.childNodes[i];
                if (node.nodeType===1){
                    if (node.tagName==="g"){
                        res.n=res.n+1;
                    }else if(node.tagName==="rect"){
                        res.x=node.x.baseVal.value+2;
                        res.y=node.y.baseVal.value+2;
                    }
                }
            }
            return res;
        };
        card.stripChildIds =function(group){
            let all = group.getElementsByTagName("*");

            for (let i=0, max=all.length; i < max; i++) {
                if (all[i].id){
                    all[i].removeAttribute("id");
                }
            }
        };

        card.set=function(group,cardGroup,vertical){
            let vspace=0;
            let hspace=0;
            let cardFrame=cardGroup.getElementsByTagName("rect")[0];
            if (vertical){
                vspace=hand.vSpace;
            }else{
                hspace=hand.hSpace;
            }
            let cards,posX,posY;
            if (group.id.charAt(0)==="p"){
                ({n:cards,x:posX,y:posY}=card.count(group));
            }else{
                let gpid="p"+group.id.substring(1,group.id.length);
                ({x:posX,y:posY}=card.count(document.getElementById(gpid)));
                ({n:cards}=card.count(group));
            }
            let newX=vspace*cards+posX-cardFrame.x.baseVal.value;
            let newY=hspace*cards+posY-cardFrame.y.baseVal.value;
            if ( cardGroup.transform.baseVal.length===1){
                cardGroup.transform.baseVal.getItem(0).setTranslate(newX,newY);
            }else{
                let matrix = document.createElementNS("http://www.w3.org/2000/svg", "svg").createSVGMatrix();
                matrix=matrix.translate(newX,newY);
                let item=cardGroup.transform.baseVal.createSVGTransformFromMatrix(matrix);
                cardGroup.transform.baseVal.appendItem(item);
            }
            group.appendChild(cardGroup);
        };
        //leave does not remove the card just shift the other cards
        //when the card is insert in another group is should move.
        card.leave=function(cardGroup,vertical){
            let group=cardGroup.parentNode;
            let cards=group.getElementsByTagName("g");
            for(let i=cards.length-1;i>=0;i--){
                if (cards[i].id===cardGroup.id){
                    break;
                }else{
                    let newX,newY;
                    if (vertical){
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e-hand.vSpace;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f;
                    }else{
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f-hand.hSpace;
                    }
                    cards[i].transform.baseVal.getItem(0).setTranslate(newX,newY);
                }
            }
        };
        card.clearGroup=function(group){
            let cards=group.getElementsByTagName("g");
            for(let i=cards.length-1;i>=0;i--){
                group.removeChild(cards[i]);
            }
        };
        card.hit=function(group,x,y){
            let res=[];
            let cards=group.getElementsByTagName("g");
            if (cards.length!==0){
                for(let i=cards.length-1;i>=0;i--){
                    let rect=cards[i].getElementsByTagName("rect")[0];
                    let y0=rect.y.baseVal.value+cards[i].transform.baseVal.getItem(0).matrix.f;
                    if (cards[i].transform.baseVal.numberOfItems===2){
                        y0=y0+cards[i].transform.baseVal.getItem(1).matrix.f;
                    }
                    let y1=y0+rect.height.baseVal.value;
                    let x0=rect.x.baseVal.value+cards[i].transform.baseVal.getItem(0).matrix.e;
                    let x1=x0+rect.width.baseVal.value;

                    let area=new Area(x0,x1,y0,y1);
                    if( area.hit(x,y)){
                        res[0]=cards[i];
                        break;
                    }
                }
            }
            return res;
        };
        hand.select=function(cardGroup){
            let matrix = document.createElementNS("http://www.w3.org/2000/svg", "svg").createSVGMatrix();
            matrix=matrix.translate(0,-20);
            let item=cardGroup.transform.baseVal.createSVGTransformFromMatrix(matrix);
            cardGroup.transform.baseVal.appendItem(item);
            hand.selected=cardGroup;
        };
        exp.hand.select=hand.select;
        exp.hand.selected=function(){
            return hand.selected;
        };
        hand.unSelect=function(){
            hand.selected.transform.baseVal.removeItem(1);
            hand.selected=null;
        };
        exp.hand.unSelect=hand.unSelect;

        flag.cardLastMoveMark=function(cardix){
            let cardGroup=document.getElementById(id.typeToName(ID_Card,cardix));
            cardGroup.getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrameLastMove;
            if (flag.cardLastMoveSelected){
                flag.cardLastMoveSelected.push(cardGroup);
            }else{
                flag.cardLastMoveSelected=[cardGroup];
            }
        };
        exp.flag.cardLastMoveMark=flag.cardLastMoveMark;
        flag.cardLastMoveUnMark=function(){
            if (flag.cardLastMoveSelected){
                for(let i=0;i<flag.cardLastMoveSelected.length;i++){
                    flag.cardLastMoveSelected[i].getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrame;
                }
                flag.cardLastMoveSelected=null;
            }
        };
        exp.flag.cardLastMoveUnMark=flag.cardLastMoveUnMark;
        flag.cardSelect=function(cardGroup){
            cardGroup.getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrameSpecial;
            flag.cardSelected=cardGroup;
        };
        exp.flag.cardSelect=flag.cardSelect;
        exp.flag.cardSelected=function(){
            return flag.cardSelected;
        };
        flag.cardUnSelect=function(){
            flag.cardSelected.getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrame;
            flag.cardSelected=null;
        };
        exp.flag.cardUnSelect=flag.cardUnSelect;

        flag.cardToDish=function(cardX){
            let cardGroup;
            if (cardX.parentNode){
                cardGroup=cardX;
            }else{
                cardGroup=document.getElementById(id.typeToName(ID_Card,cardX));
            }
            card.leave(cardGroup,false);
            card.moveToDish(cardGroup);
        };
        exp.flag.cardToDish=flag.cardToDish;

        flag.cardToFlag=function(cardX,flagX,player){
            let cardGroup;
            let cardNo;
            if (cardX.parentNode){
                cardGroup=cardX;
                cardNo=id.fromName(cardGroup.id).no;
            }else{
                cardGroup=document.getElementById(id.typeToName(ID_Card,cardX));
                cardNo=cardX;
            }
            let flagGroup;
            if (flagX.parentNode){
                flagGroup=flagX;
            }else{
                if (cardNo===TC_Mud||cardNo===TC_Fog){
                    flagGroup=document.getElementById(id.typeToName(ID_FlagTac,flagX,player));
                }else{
                    flagGroup=document.getElementById(id.typeToName(ID_FlagTroop,flagX,player));
                }
            }
            card.leave(cardGroup,false);
            card.set(flagGroup,cardGroup,false);
        };
        exp.flag.cardToFlag=flag.cardToFlag;

        flag.cardToFlagPlayer=function(flagGroup){
            let cardGroup=flag.cardSelected;
            flag.cardUnSelect();
            flag.cardToFlag(cardGroup,flagGroup,true);
        };
        exp.flag.cardToFlagPlayer=flag.cardToFlagPlayer;

        cone.pos=function(coneX,pos){
            let coneCircle;
            if (coneX.parentNode){
                coneCircle=coneX;
            }else{
                coneCircle=document.getElementById(id.typeToName(ID_Cone,coneX));
            }
            let center=350;
            let move=262;
            let newY;
            switch(pos){
            case 0:
                newY=350-262;
                break;
            case 1:
                newY=350;
                break;
            case 2:
                newY=350+262;
                break;
            }
            coneCircle.cy.baseVal.value=newY ;
        };
        exp.cone.pos=cone.pos;
        cone.clear=function(){
            for (let i =1;i<10;i++){
                cone.pos(i,1);
            }
        };

        exp.card.tacName=card.tacName;
        function clear(){
            card.clear();
            cone.clear();
            hand.selected=null;
            flag.cardSelected=null;
        }
        exp.clear=clear;
        card.colorName=function(cardNo){
            return COLORS_Names[Math.floor((cardNo-1)/10)];
        };
        exp.card.colorName=card.colorName;
        card.troopValue =function(cardNo){
           let no=""+cardNo%10;
            if (no==="0"){
                no="10";
            }
            return no;
        };
        exp.card.troopValue=card.troopValue;
        card.tacName =function(cardNo){
            let nameTxt;
            switch(cardNo){
            case TC_Traitor:
                nameTxt="Traitor";
                break;
            case TC_Alexander:
                nameTxt="Alexander";
                break;
            case TC_Darius:
                nameTxt="Darius";
                break;
            case TC_123:
                nameTxt="123";
                break;
            case TC_8:
                nameTxt="8";
                break;
            case TC_Deserter:
                nameTxt="Deserter";
                break;
            case TC_Redeploy:
                nameTxt="Redeploy";
                break;
            case TC_Scout:
                nameTxt="Scout";
                break;
            case TC_Mud:
                nameTxt="Mud";
                break;
            case TC_Fog:
                nameTxt="Fog";
                break;
            }
            return nameTxt;
        };
        let backTroop= document.getElementById("backTroopGroup").cloneNode(true);
        let backTroopRect=document.getElementById("backTroopTopRect");
        let backTroopColor=backTroopRect.style.stroke;
        let backTac= document.getElementById("backTacGroup").cloneNode(true);
        let backTacRect=document.getElementById("backTacTopRect");
        let backTacColor=backTacRect.style.stroke;
        let lTroop1=document.getElementById("troop1Group");
        let lTroop10=document.getElementById("troop10Group");
        let troop1=lTroop1.cloneNode(true);
        let troop10=lTroop10.cloneNode(true);
        let pDishTroopGroup=document.getElementById("pDishTroopGroup");
        pDishTroopGroup.removeChild(lTroop10);
        pDishTroopGroup.removeChild(lTroop1);

        let lTac=document.getElementById("tacGroup");
        let tac=lTac.cloneNode(true);
        let pDishTacGroup=document.getElementById("pDishTacGroup");
        pDishTacGroup.removeChild(lTac);


        hand.createCard=function(cardNo){
            let cardGroup;
            let text;
            let color;
            if (card.isTac(cardNo)){
                text=card.tacName(cardNo);
                cardGroup=tac.cloneNode(true);
            }else{
                text=card.troopValue(cardNo);
                color=COLORS_Troops [Math.floor((cardNo-1)/10)];
                if (text==="10"){
                    cardGroup=troop10.cloneNode(true);
                }else{
                    cardGroup=troop1.cloneNode(true);
                }
            }
            let texts=cardGroup.getElementsByTagName("tspan");
            texts[0].textContent=text;
            texts[1].textContent=text;
            if (color){
                cardGroup.getElementsByTagName("rect")[0].style.fill=color;
            }

            cardGroup.id=id.typeToName(ID_Card,cardNo,false);
            card.stripChildIds(cardGroup);
            return cardGroup;
        };
        hand.createBack=function(troop){
            let cardGroup;
            if (troop){
                cardGroup=backTroop.cloneNode(true);
            }else{
                cardGroup=backTac.cloneNode(true);
            }
            cardGroup.removeAttribute("id");
            card.stripChildIds(cardGroup);
            return cardGroup;
        };
        let pHandGroup=document.getElementById("pHandGroup");
        let deckTacTspan=document.getElementById("deckTacTspan");
        let deckTroopTspan=document.getElementById("deckTroopTspan");

        hand.drawPlayer=function(cardNo){
            if (hand.selected){
                hand.unSelect();
            }
            let cardGroup=hand.createCard(cardNo);
            card.set(pHandGroup,cardGroup,true);
            hand.select(cardGroup);
            if (card.isTac(cardNo)){
                deckTacTspan.textContent=""+deckTacTspan.textContent-1;
            }else{
                deckTroopTspan.textContent=""+deckTroopTspan.textContent-1;
            }
        };
        exp.hand.drawPlayer=hand.drawPlayer;
        let oHandGroup=document.getElementById("oHandGroup");

        hand.drawOpp=function(troop){
            let cardGroup=hand.createBack(troop);
            card.set(oHandGroup,cardGroup,true);
            if (troop){
                deckTroopTspan.textContent=""+deckTroopTspan.textContent-1;
            }else{
                deckTacTspan.textContent=""+deckTacTspan.textContent-1;
            }
        };
        exp.hand.drawOpp=hand.drawOpp;
        hand.countPlayer= function(){
            return card.count(pHandGroup);
        };

        hand.move=function(toCard,before){
            let fromCard=hand.selected;
            hand.unSelect();
            let cards=pHandGroup.getElementsByTagName("g");
            let moves=0;
            function moveCard(cc,m){
                moves=moves+m;
                let newX=cc.transform.baseVal.getItem(0).matrix.e+m;
                let newY=cc.transform.baseVal.getItem(0).matrix.f;
                cc.transform.baseVal.getItem(0).setTranslate(newX,newY);
            }
            for (let i=0,move=0;i<cards.length;i++){
                if (cards[i].id===toCard.id){
                    if(move===0){
                        move=hand.vSpace;
                        if (before){
                            moveCard(cards[i],move);
                        }
                    }else{
                        if (!before){
                            moveCard(cards[i],move);
                        }
                        break;
                    }
                }else if(cards[i].id===fromCard.id){
                    if (move===0){
                        move=-hand.vSpace;
                    }else{
                        break;
                    }
                }else{
                    if (move!==0){
                        moveCard(cards[i],move); 
                    }
                }
            }
            if(moves!==0){
                moveCard(fromCard,-moves);
                if (before){
                    pHandGroup.insertBefore(fromCard,toCard);
                }else{
                    if (toCard.nextElementSibling){
                        pHandGroup.insertBefore(fromCard,toCard.nextElementSibling);
                    }else{
                        pHandGroup.appendChild(fromCard);
                    }
                }
            }
        };
        exp.hand.move=hand.move;

        hand.moveToDishPlayer=function(){
            let c=hand.selected;
            hand.unSelect();
            card.leave(c,true);
            card.moveToDish(c,true);
        };
        exp.hand.moveToDishPlayer=hand.moveToDishPlayer;
        hand.removeOpp=function(cardNo){
            let cards=oHandGroup.getElementsByTagName("g");
            let color;
            if (card.isTac(cardNo)){
                color=backTacColor;
            }else{
                color=backTroopColor;
            }
            for(let i=cards.length-1;i>=0;i--){
                if(cards[i].getElementsByTagName("rect")[1].style.stroke===color){
                    oHandGroup.removeChild(cards[i]);
                    break;
                }else{
                    let newX,newY;
                    let cc=cards[i];
                    newX=cards[i].transform.baseVal.getItem(0).matrix.e-hand.vSpace;
                    newY=cards[i].transform.baseVal.getItem(0).matrix.f;
                    cards[i].transform.baseVal.getItem(0).setTranslate(newX,newY);
                }
            }
        };

        hand.moveToDishOpp=function(cardNo){
            hand.removeOpp(cardNo);
            let c=hand.createCard(cardNo);
            card.moveToDish(c,false);
        };
        exp.hand.moveToDishOpp=hand.moveToDishOpp;
        let oDishTacGroup=document.getElementById("oDishTacGroup");
        let oDishTroopGroup=document.getElementById("oDishTroopGroup");
        card.moveToDish=function(cardGroup,player){
            let v=id.fromName(cardGroup.id).no;
            if(player===undefined){//work only for flag
                player=id.fromName(cardGroup.parentNode.id).player;
            }
            let dishGroup;
            if (player){
                if(card.isTac(v)){
                    dishGroup=pDishTacGroup;
                }else{
                    dishGroup=pDishTroopGroup;
                }
            }else{
                if(card.isTac(v)){
                    dishGroup=oDishTacGroup;
                }else{
                    dishGroup=oDishTroopGroup;
                }
            }
            card.set(dishGroup,cardGroup,false);
        };
        card.clear=function(){
            card.clearGroup(oDishTroopGroup);
            card.clearGroup(oDishTacGroup);
            card.clearGroup(pDishTroopGroup);
            card.clearGroup(pDishTacGroup);
            for(let i=1;i<10;i++){
                let idName= id.typeToName(ID_FlagTroop,i,true);
                card.clearGroup(document.getElementById(idName));
                idName= id.typeToName(ID_FlagTroop,i,false);
                card.clearGroup(document.getElementById(idName));
                idName= id.typeToName(ID_FlagTac,i,true);
                card.clearGroup(document.getElementById(idName));
                idName= id.typeToName(ID_FlagTac,i,false);
                card.clearGroup(document.getElementById(idName));
            }
            card.clearGroup(pHandGroup);
            card.clearGroup(oHandGroup);
            deckTacTspan.textContent=""+TAC_NO;
            deckTroopTspan.textContent=""+TROOP_NO;
        };

        hand.moveToFlagPlayer=function(flagX){
            let c=hand.selected;
            let cardix=id.fromName(c.id).no;
            let flagGroup;
            if (flagX.parentNode){
                let flagIdObj=id.fromName(flagX.id);
                if (cardix===TC_Mud||cardix===TC_Fog){
                    if(flagIdObj.type===ID_FlagTac){
                        flagGroup=flagX;
                    }else{
                        flagGroup=document.getElementById(id.typeToName(ID_FlagTac,flagIdObj.no,true));
                    }
                }else{
                    if(flagIdObj.type===ID_FlagTroop){
                        flagGroup=flagX;
                    }else{
                        flagGroup=document.getElementById(id.typeToName(ID_FlagTroop,flagIdObj.no,true));
                    }
                }
            }else{
                if (cardix===TC_Mud||cardix===TC_Fog){
                    flagGroup=document.getElementById(id.typeToName(ID_FlagTac,flagX,true));
                }else{
                    flagGroup=document.getElementById(id.typeToName(ID_FlagTroop,flagX,true));
                }
            }
            hand.unSelect();
            card.leave(c,true);
            card.set(flagGroup,c,false);
        };
        exp.hand.moveToFlagPlayer=hand.moveToFlagPlayer;

        hand.moveToFlagOpp=function(cardNo,flagNo){
            let flagGroup;
            if (cardNo===TC_Mud||cardNo===TC_Fog){
                flagGroup=document.getElementById(id.typeToName(ID_FlagTac,flagNo,false));
            }else{
                flagGroup=document.getElementById(id.typeToName(ID_FlagTroop,flagNo,false));
            }
            hand.removeOpp(cardNo);
            let c=hand.createCard(cardNo);
            card.set(flagGroup,c,false);
        };
        exp.hand.moveToFlagOpp=hand.moveToFlagOpp;
        hand.removePlayer=function(cardGroup){
            cardGroup.parentNode.removeChild(cardGroup);
        };

        hand.moveToDeckPlayer=function(){
            let cc=hand.selected;
            hand.unSelect();
            if (card.isTac(id.fromName(cc.id).no)){
                deckTacTspan.textContent=""+(parseInt(deckTacTspan.textContent)+1);
            }else{
                deckTroopTspan.textContent=""+(parseInt(deckTroopTspan.textContent)+1);
            }
            card.leave(cc,true);
            hand.removePlayer(cc,true);
        };
        exp.hand.moveToDeckPlayer=hand.moveToDeckPlayer;

        hand.moveToDeckOpp=function(tac){
            if (tac){
                hand.removeOpp(TROOP_NO+1);
                deckTacTspan.textContent=""+(parseInt(deckTacTspan.textContent)+1);
            }else{
                hand.removeOpp(TROOP_NO);
                deckTroopTspan.textContent=""+(parseInt(deckTroopTspan.textContent)+1);
            }
        };
        exp.hand.moveToDeckOpp=hand.moveToDeckOpp;
        let battlelineSvg=document.getElementById("battlelineSvg");
        click.zone={};
        click.hitElems=function(event){
            let m = battlelineSvg.getScreenCTM();
            let point = battlelineSvg.createSVGPoint();
            point.x = event.clientX;
            point.y = event.clientY;
            point = point.matrixTransform(m.inverse());

            let res=click.zone.handHit(point.x,point.y);
            if (res.length===0){
                res=click.zone.deckHit(point.x,point.y);
            }
            if(res.length===0){
                res=click.zone.flagHit(point.x,point.y,true);
            }
            if(res.length===0){
                res=click.zone.flagHit(point.x,point.y,false);
            }
            if(res.length===0){
                res=click.zone.coneHit(point.x,point.y);
            }
            if(res.length===0){
                res=click.zone.dishHit(point.x,point.y);
            }
            return res;
        };
        click.deaFlagSet=new Set();
        click.deactivateFlag=function(flagNo){
            click.deaFlagSet.add(flagNo);
        };
        click.resetDeactivate=function(){
            click.deaFlagSet.clear();
        };
        let pHandRec=pHandGroup.getElementsByTagName("rect")[0];
        let frameStroke=pHandRec.style.strokeWidth/2;
        let pHandArea=new Area(pHandRec.x.baseVal.value+frameStroke,
                               pHandRec.x.baseVal.value-frameStroke+pHandRec.width.baseVal.value,
                               pHandRec.y.baseVal.value+frameStroke-hand.hSpace,
                               pHandRec.y.baseVal.value-frameStroke+pHandRec.height.baseVal.value
                              );
        click.zone.handHit=function(x,y){
            let res=[];
            if (pHandArea.hit(x,y)){
                res=card.hit(pHandGroup,x,y);
            }
            return res;
        };
        let pF1Troop=document.getElementById("pF1TroopGroup");
        let pF1TroopRec=pF1Troop.getElementsByTagName("rect")[0];
        let pF9Tac=document.getElementById("pF9TacGroup");
        let pF9TacRec=pF9Tac.getElementsByTagName("rect")[0];
        let pFlagArea=new Area(pF1TroopRec.x.baseVal.value+frameStroke,
                               pF9TacRec.x.baseVal.value-frameStroke+pF9TacRec.width.baseVal.value,
                               pF1TroopRec.y.baseVal.value+frameStroke,
                               pF9TacRec.y.baseVal.value-frameStroke+pF9TacRec.height.baseVal.value
                              );
        let oppMatrix=document.getElementById("oppGroup").transform.baseVal.getItem(0).matrix;
        click.zone.flagHit=function(x,y,player){
            let res=[];
            if (!player){
                let point = battlelineSvg.createSVGPoint();
                point.x = x;
                point.y = y;
                point = point.matrixTransform(oppMatrix.inverse());
                x=point.x;
                y=point.y;
            }
            function hitGroup(group,x,y){
                let hit=false;
                let rect=group.getElementsByTagName("rect")[0];
                if (!player){
                    let m=group.transform.baseVal.getItem(0).matrix;
                    let point = battlelineSvg.createSVGPoint();
                    point.x = x;
                    point.y = y;
                    point = point.matrixTransform(m.inverse());
                    x=point.x;
                    y=point.y;
                }
                let y0=rect.y.baseVal.value;
                let y1=rect.y.baseVal.value+rect.height.baseVal.value;
                let x0=rect.x.baseVal.value;
                let x1=rect.x.baseVal.value+rect.width.baseVal.value;
                let area=new Area(x0,x1,y0,y1);
                if( area.hit(x,y)){
                    hit=true;
                }
                return hit;
            }
            if (pFlagArea.hit(x,y)){
                for(let i=1;i<10;i++){
                    let troopGroup=document.getElementById(id.typeToName(ID_FlagTroop,i,player));
                    res=card.hit(troopGroup,x,y);
                    if (res.length===0){
                        let tacGroup=document.getElementById(id.typeToName(ID_FlagTac,i,player));
                        res=card.hit(tacGroup,x,y);
                        if (res.length>0){
                            res[res.length]=tacGroup;
                            break;
                        }
                        if (hitGroup(troopGroup,x,y)){
                            res[res.length]=troopGroup;
                            break;
                        }
                        if (hitGroup(tacGroup,x,y)){
                            res[res.length]=tacGroup;
                            break;
                        }
                    }else{
                        res[res.length]=troopGroup;
                        break;
                    }
                }
            }
            return res;
        };
        let deckTacGroup=document.getElementById("deckTacGroup");
        let deckTroopGroup=document.getElementById("deckTroopGroup");
        click.zone.deckHit=function(x,y){
            let res=[];
            let troopArea=new Area(backTroopRect.x.baseVal.value,
                                   backTroopRect.x.baseVal.value+backTroopRect.width.baseVal.value,
                                   backTroopRect.y.baseVal.value,
                                   backTroopRect.y.baseVal.value+backTroopRect.height.baseVal.value
                                  );
            if(troopArea.hit(x,y)){
                res[0]=deckTroopGroup;
            }else{
                let tacArea=new Area(backTacRect.x.baseVal.value,
                                     backTacRect.x.baseVal.value+backTacRect.width.baseVal.value,
                                     backTacRect.y.baseVal.value,
                                     backTacRect.y.baseVal.value+backTacRect.height.baseVal.value
                                    );
                if(tacArea.hit(x,y)){
                    res[0]=deckTacGroup;
                }
            }
            return res;
        };
        let pDishTacRect=pDishTacGroup.getElementsByTagName("rect")[0];
        let pDishTroopRect=pDishTroopGroup.getElementsByTagName("rect")[0];
        click.zone.dishHit=function(x,y){
            let res=[];
            let troopArea=new Area(pDishTroopRect.x.baseVal.value,
                                   pDishTroopRect.x.baseVal.value+pDishTroopRect.width.baseVal.value,
                                   pDishTroopRect.y.baseVal.value,
                                   pDishTroopRect.y.baseVal.value+pDishTroopRect.height.baseVal.value
                                  );
            if(troopArea.hit(x,y)){
                res[0]=pDishTroopGroup;
            }else{
                let tacArea=new Area(pDishTacRect.x.baseVal.value,
                                     pDishTacRect.x.baseVal.value+pDishTacRect.width.baseVal.value,
                                     pDishTacRect.y.baseVal.value,
                                     pDishTacRect.y.baseVal.value+pDishTacRect.height.baseVal.value
                                    );
                if(tacArea.hit(x,y)){
                    res[0]=pDishTacGroup;
                }
            }
            return res;
        };
        let firstCone=document.getElementById("k1Path");
        let lastCone=document.getElementById("k9Path");
        let coneR=firstCone.r.baseVal.value+parseFloat(firstCone.style.strokeWidth);
        let coneArea=new Area(firstCone.cx.baseVal.value-coneR,
                              lastCone.cx.baseVal.value+coneR,
                              firstCone.cy.baseVal.value-coneR,
                              lastCone.cy.baseVal.value+coneR
                             );
        click.zone.coneHit=function(x,y){
            let res=[];
            if(coneArea.hit(x,y)){
                for (let i=1;i<10;i++){
                    if(!click.deaFlagSet.has(i)){
                        let coneElm= document.getElementById(id.typeToName(ID_Cone,i));
                        let r=coneElm.r.baseVal.value+parseFloat(coneElm.style.strokeWidth);
                        let coneArea=new Area(coneElm.cx.baseVal.value-r,
                                              coneElm.cx.baseVal.value+r,
                                              coneElm.cy.baseVal.value-r,
                                              coneElm.cy.baseVal.value+r
                                             );
                        if (coneArea.hit(x,y)){
                            res[0]=coneElm;
                        }
                    }
                }
            }
            return res;
        };
        function itemClicked(elems,centerClick){
        };
        let clickFailAudio=document.getElementById("fail-click-audio");
        exp.itemClicked=itemClicked;
        battlelineSvg.onclick=function (event){
            let elms=click.hitElems(event);
            let centerClick=event.which===2;//right click is context menu
            let used=false;
            if (elms.length>0){
                used=exp.itemClicked(elms,centerClick);
            }
            if (!used){
                clickFailAudio.play();
            }
        };
        return exp;
    }

    function getCookies(document){
        let res=new Map();
        let cookies=document.cookie;
        if(cookies!==""){
            let list = cookies.split("; ");
            for(let i=0;i<list.length;i++){
                let cookie=list[i];
                let p=cookie.indexOf("=");
                let name=cookie.substring(0,p);
                let value=cookie.substring(p+1);
                value=decodeURIComponent(value);
                res[name]=value;
            }
        }
        return res;
    }
    function createGame(document,ws,svg,msg){
        const TURN_FLAG = 0;
	      const TURN_HAND = 1;
	      //TURN_SCOUT2 player picks second of tree scout cards.
	      const TURN_SCOUT2 = 2;
	      //TURN_SCOUT2 player picks last of tree scout cards.
	      const TURN_SCOUT1 = 3;
	      //TURN_SCOUTR player return 3 cards to decks.
	      const TURN_SCOUTR = 4;
	      const TURN_DECK   = 5;
	      const TURN_FINISH = 6;
	      const TURN_QUIT   = 7;

	      const DECK_TAC   = 1;
	      const DECK_TROOP = 2;


        let exp={};
        let turn={};
        turn.current=null;
        let cone={};
        cone.clickedixs=new Set();
        cone.validixs=new Set();

        let scoutReturnTacixs=[];
        let scoutReturnTroopixs=[];

        function isPlaying(){
            return turn.current!==null;
        }
        exp.isPlaying=isPlaying;
        function showMove(moveView){
            cone.validixs.clear();
            let move=moveView.Move;
            if (moveView.Mover){
                if (move.JsonType==="MoveClaimView"){
                    if (move.Claimed.length!==move.Claim.length){
                        for(let i=0;i<move.Claim.length;i++){
                            let found=false;
                            for(let j=0;j<move.Claimed.length;j++){
                                if (move.Claim[i]===move.Claimed[j]){
                                    found=true;
                                    break;
                                }
                            }
                            if (!found){
                                svg.cone.pos(move.Claim[i]+1,1);//reset cone
                            }
                        }
                        function exTxt(cardixs){
                            let txt="";
                            if (cardixs){//we do not always calculate a example
                                for(let i=0;i<cardixs.length;i++){
                                    if (cardixs[i]!==0){
                                        if (svg.card.isTac(cardixs[i])){
                                            txt=txt+svg.card.tacName;
                                        }else{
                                            txt=txt+svg.card.colorName(cardixs[i]);
                                            txt=txt+" ";
                                            txt=txt+svg.card.troopValue(cardixs[i]);
                                        }
                                        txt=txt+", ";
                                    }
                                }
                                txt=txt.substr(0,txt.length-2);
                            }

                            return txt;
                        }

                        let txt="";
                        let ex="";
                        for(let flagix in move.FailMap){
                            txt=txt+"Claim failed for flag: "+(parseInt(flagix)+1);
                            txt=txt+"\n";
                            ex=exTxt(move.FailMap[flagix]);
                            if (ex!==""){
                                txt=txt+"For example: "+ex;
                                txt=txt+"\n";
                            }
                        }
                        msg.recieved({Message:txt.substr(0,txt.length-1)});
                    }
                    if (move.Win){
                        msg.recieved({Message:"Congratulation you won the game."});
                    }
                }else if (moveView.DeltCardix!==0){
                    if (svg.hand.selected()){
                        svg.hand.unSelect();
                    }
                    svg.hand.drawPlayer(moveView.DeltCardix);
                }else if(move.JsonType==="MoveQuit"){
                    msg.recieved({Message:"You have lost the game by giving up."});
                }else if(move.JsonType==="MoveRedeployView"){
                    if(move.RedeployDishixs){
                        for(let i=0;i<move.RedeployDishixs.length;i++){
                            svg.flag.cardToDish(move.RedeployDishixs[i]);
                        }
                    }
                }else if(move.JsonType==="MoveDeserterView"){
                    if (move.Dishixs){
                        for(let i=0;i<move.Dishixs.length;i++){
                            svg.flag.cardToDish(move.Dishixs[i]);
                        }
                    }
                }
            }else{//Opponent move
                //Move bat.Move and Card int
                switch (move.JsonType){
                case "MoveInit":
                    for (let i=0;i<move.Hand.length;i++){
                        svg.hand.drawPlayer(move.Hand[i]);
                        svg.hand.drawOpp(true);
                    }
                    break;
                case "MoveInitPos":
                    let pos=move.Pos;
                    for (let i=0;i<pos.Flags.length;i++){
                        let flag=pos.Flags[i];
                        if (flag.OppFlag){
                            svg.cone.pos(i+1,0);
                        }else if(flag.NeuFlag){
                            svg.cone.pos(i+1,1);
                        }else if(flag.PlayFlag){
                            svg.cone.pos(i+1,2);
                        }
                        for (let j=0;j<flag.OppTroops.length;j++){
                            let troop=true;
                            if (svg.card.isTac(flag.OppTroops[j])){
                                troop=false;
                            }
                            svg.hand.drawOpp(troop);
                            svg.hand.moveToFlagOpp(flag.OppTroops[j],i+1);
                        }
                        for (let j=0;j<flag.OppEnvs.length;j++){
                            svg.hand.drawOpp(false);
                            svg.hand.moveToFlagOpp(flag.OppEnvs[j],i+1);
                        }
                        for (let j=0;j<flag.PlayTroops.length;j++){
                            svg.hand.drawPlayer(flag.PlayTroops[j]);
                            svg.hand.moveToFlagPlayer(i+1);
                        }
                        for (let j=0;j<flag.PlayEnvs.length;j++){
                            svg.hand.drawPlayer(flag.PlayEnvs[j]);
                            svg.hand.moveToFlagPlayer(i+1);
                        }
                    }
                    for (let i=0;i<pos.OppDishTroops.length;i++){
                        let cardNo=pos.OppDishTroops[i];
                        svg.hand.drawOpp(true);
                        svg.hand.moveToDishOpp(cardNo);
                    }
                    for (let i=0;i<pos.OppDishTacs.length;i++){
                        let cardNo=pos.OppDishTacs[i];
                        svg.hand.drawOpp(false);
                        svg.hand.moveToDishOpp(cardNo);
                    }
                    for (let i=0;i<pos.DishTroops.length;i++){
                        let cardNo=pos.DishTroops[i];
                        svg.hand.drawPlayer(cardNo);
                        svg.hand.moveToDishPlayer();
                    }
                    for (let i=0;i<pos.DishTacs.length;i++){
                        let cardNo=pos.DishTacs[i];
                        svg.hand.drawPlayer(cardNo);
                        svg.hand.moveToDishPlayer();
                    }
                    for (let i=0;i<pos.OppHand.length;i++){
                        svg.hand.drawOpp(pos.OppHand[i]);
                    }
                    for (let i=0;i<pos.Hand.length;i++){
                        let cardNo=pos.Hand[i];
                        svg.hand.drawPlayer(cardNo);
                        svg.hand.unSelect();
                    }
                    break;
                case "MoveCardFlag":
                    svg.hand.moveToFlagOpp(moveView.MoveCardix,move.Flagix+1);
                    svg.flag.cardLastMoveUnMark();
                    svg.flag.cardLastMoveMark(moveView.MoveCardix);
                    break;
                case "MoveDeck":
                    svg.hand.drawOpp(move.Deck===DECK_TROOP);
                    if (moveView.MoveCardix===svg.card.TC_Scout){
                        svg.hand.moveToDishOpp(svg.card.TC_Scout);
                    }
                    break;
                case "MoveClaimView":
                    if (move.Claimed.length>0){
                        for(let i=0;i<move.Claimed.length;i++){
                            svg.cone.pos(move.Claimed[i]+1,0);
                        }
                        if (move.Win){
                            msg.recieved({Message:"Sorry you you lost the game."});
                        }
                    }
                    break;
                case "MoveDeserterView":
                    svg.hand.moveToDishOpp(moveView.MoveCardix);
                    svg.flag.cardToDish(move.Move.Card);
                    svg.flag.cardLastMoveUnMark();
                    svg.flag.cardLastMoveMark(moveView.MoveCardix);
                    svg.flag.cardLastMoveMark(move.Move.Card);
                    if (move.Dishixs){
                        for(let i=0;i<move.Dishixs.length;i++){
                            svg.flag.cardToDish(move.Dishixs[i]);
                            svg.flag.cardLastMoveMark(move.Dishixs[i]);
                        }
                    }
                    break;
                case "MoveScoutReturnView":
                    if (move.Tac>0){
                        for (let i=0;i<move.Tac;i++){
                            svg.hand.moveToDeckOpp(true);
                        }
                    }
                    if (move.Troop>0){
                        for (let i=0;i<move.Troop;i++){
                            svg.hand.moveToDeckOpp(false);
                        }
                    }
                    svg.flag.cardLastMoveUnMark();
                    svg.flag.cardLastMoveMark(moveView.MoveCardix);

                    msg.recieved({Message:"Opponent return "+move.Tac+" tactic cards and "+move.Troop+" troop cards."});
                    break;
                case "MoveTraitor":
                    svg.hand.moveToDishOpp(moveView.MoveCardix);
                    svg.flag.cardToFlag(move.OutCard,move.InFlag+1,false);

                    svg.flag.cardLastMoveUnMark();
                    svg.flag.cardLastMoveMark(moveView.MoveCardix);
                    svg.flag.cardLastMoveMark(move.OutCard);
                    break;
                case "MoveRedeployView":
                    svg.hand.moveToDishOpp(moveView.MoveCardix);
                    svg.flag.cardLastMoveUnMark();
                    svg.flag.cardLastMoveMark(moveView.MoveCardix);
                    svg.flag.cardLastMoveMark(move.OutCard);
                    if (move.Move.InFlag>=0){
                        svg.flag.cardToFlag(move.Move.OutCard,move.Move.InFlag+1,false);
                    }else{
                        svg.flag.cardToDish(move.Move.OutCard);
                    }
                    if(move.RedeployDishixs){
                        for(let i=0;i<move.RedeployDishixs.length;i++){
                            svg.flag.cardToDish(move.RedeployDishixs[i]);
                            svg.flag.cardLastMoveMark(move.RedeployDishixs[i]);
                        }
                    }

                    break;
                case "MovePass":
                    msg.recieved({Message:"Your opponent chose not to play a card."});
                    break;
                case "MoveQuit":
                    msg.recieved({Message:"Congratulation your opponent have given up."});
                    break;
                case "MoveSave":
                    msg.recieved({Message:"Oppent have requested the game to be saved to be continued"});
                    break;
                default:
                    console.log("Unsupported move: "+move.JsonType);

                }
            }
        };
        cone.clear=function(){
            cone.clickedixs.clear();
            cone.validixs.clear();
        };
        function clear(){
            cone.clear();
            turn.clear();
            svg.clear();
            scoutReturnTroopixs=[];
            scoutReturnTacixs=[];
        };
        exp.clear=clear;

        function onClickedCard(clickedFlagElm,clickedCardElm,clickedDishElm){
            let used=false;
            function moveToFlag(cardix,flagElm){
                let used=false;
                let flagIdObj=svg.id.fromName(flagElm.id);
                let flagNo=flagIdObj.no;
                if (flagIdObj.player){
                    let moves=turn.getHandMove(cardix);
                    if (moves){
                        for(let i=0;i<moves.length;i++){
                            if (moves[i].Flagix===flagNo-1){
                                svg.hand.moveToFlagPlayer(flagElm);
                                turn.isMyTurn=false;
                                ws.actionBuilder(ws.ACT_MOVE).move(cardix,i).send();
                                used=true;
                                break;
                            }
                        }
                    }
                }
                return used;
            }
            if(turn.isMyTurn&&svg.hand.selected()!==null&&turn.getState()===TURN_HAND){
                let player;
                let clickedFlagix;
                if (clickedFlagElm){
                    let flagIdObj=svg.id.fromName(clickedFlagElm.id);
                    player=flagIdObj.player;
                    clickedFlagix=flagIdObj.no-1;
                }else{
                    player=svg.id.fromName(clickedDishElm.id).player;
                    clickedFlagix=-1;
                }
                let selectedHandCardix=svg.id.fromName(svg.hand.selected().id).no;
                if (svg.card.isTac(selectedHandCardix)){
                    switch (selectedHandCardix){
                    case svg.card.TC_123:
                    case svg.card.TC_8:
                    case svg.card.TC_Fog:
                    case svg.card.TC_Mud:
                    case svg.card.TC_Alexander:
                    case svg.card.TC_Darius:
                       used= moveToFlag(selectedHandCardix,clickedFlagElm);
                        break;
                    case svg.card.TC_Deserter:
                        if (clickedCardElm&&!player){
                            let clickedCardix=svg.id.fromName(clickedCardElm.id).no;
                            let dmoves=turn.getHandMove(selectedHandCardix);
                            for(let i=0;i<dmoves.length;i++){
                                if (dmoves[i].Card===clickedCardix){
                                    svg.hand.moveToDishPlayer();
                                    svg.flag.cardToDish(clickedCardix);
                                    turn.isMyTurn=false;
                                    ws.actionBuilder(ws.ACT_MOVE).move(selectedHandCardix,i).send();
                                    used=true;
                                    break;
                                }
                            }
                        }
                        break;
                    case svg.card.TC_Redeploy:
                        if (player){
                            if (!svg.flag.cardSelected()){
                                if (clickedCardElm){
                                    svg.flag.cardSelect(clickedCardElm);
                                    used=true;
                                }
                            }else{
                                if (clickedCardElm &&svg.flag.cardSelected().id===clickedCardElm.id){
                                    svg.flag.cardUnSelect();
                                    used=true;
                                }else{
                                    let selectedFlagCardix=svg.id.fromName(svg.flag.cardSelected().id).no;
                                    let rmoves=turn.getHandMove(selectedHandCardix);
                                    for(let i=0;i<rmoves.length;i++){
                                        if (rmoves[i].OutCard===selectedFlagCardix&&rmoves[i].InFlag===clickedFlagix){
                                            svg.hand.moveToDishPlayer();
                                            turn.isMyTurn=false;
                                            ws.actionBuilder(ws.ACT_MOVE).move(selectedHandCardix,i).send();
                                            if(clickedFlagix!==-1){
                                                svg.flag.cardToFlagPlayer(clickedFlagElm);
                                            }else{
                                                svg.flag.cardUnSelect();
                                                svg.flag.cardToDish(selectedFlagCardix);
                                            }
                                            used=true;
                                            break;
                                        }
                                    }
                                }
                            }
                        }
                        break;
                    case svg.card.TC_Traitor:
                        if (player){
                            if (svg.flag.cardSelected()){
                                let selectedFlagCardix=svg.id.fromName(svg.flag.cardSelected().id).no;
                                let tmoves=turn.getHandMove(selectedHandCardix);
                                for(let i=0;i<tmoves.length;i++){
                                    if (tmoves[i].OutCard===selectedFlagCardix&&tmoves[i].InFlag===clickedFlagix){
                                        svg.hand.moveToDishPlayer();
                                        turn.isMyTurn=false;
                                        ws.actionBuilder(ws.ACT_MOVE).move(selectedHandCardix,i).send();
                                        svg.flag.cardToFlagPlayer(clickedFlagElm);
                                        used=true;
                                        break;
                                    }
                                }
                            }
                        }else{//clicked on opp flag
                            if (!svg.flag.cardSelected()){
                                if (clickedCardElm && !svg.card.isTac(svg.id.fromName(clickedCardElm.id).no)){
                                    svg.flag.cardSelect(clickedCardElm);
                                    used=true;
                                }
                            }else{
                                if (clickedCardElm &&svg.flag.cardSelected().id===clickedCardElm.id){
                                    svg.flag.cardUnSelect();
                                    used=true;
                                }
                            }
                        }
                        break;
                    }
                }else{//TROOP
                    used=moveToFlag(selectedHandCardix,clickedFlagElm);
                }
            }
            return used;
        };
        function onClickedDeck(deckElm,idType){
            let used=false;
            if(turn.isMyTurn){
                let deck;
                if (idType===svg.id.ID_DeckTac){
                    deck=DECK_TAC;
                }else{
                    deck=DECK_TROOP;
                }
                if (turn.getState()===TURN_SCOUT1||
                    turn.getState()===TURN_SCOUT2||
                    turn.getState()===TURN_DECK){
                    let moves=turn.getMoves();
                    for(let i=0;i<moves.length;i++){
                        if (moves[i].Deck===deck){
                            turn.isMyTurn=false;
                            ws.actionBuilder(ws.ACT_MOVE).move(0,i).send();
                            used=true;
                            break;
                        }
                    }
                }else if (turn.getState()===TURN_HAND && svg.hand.selected() &&
                          svg.id.fromName(svg.hand.selected().id).no===svg.card.TC_Scout){
                    let moves=turn.getHandMove(svg.card.TC_Scout);
                    for(let i=0;i<moves.length;i++){
                        if (moves[i].Deck===deck){
                            svg.hand.moveToDishPlayer();
                            turn.isMyTurn=false;
                            ws.actionBuilder(ws.ACT_MOVE).move(svg.card.TC_Scout,i).send();
                            used=true;
                            break;
                        }
                    }
                }else if(turn.getState()===TURN_SCOUTR && svg.hand.selected()){
                    let selectedHandCardix=svg.id.fromName(svg.hand.selected().id).no;
                    let handCount;
                    if(svg.card.isTac(selectedHandCardix)){
                        if(deck===DECK_TAC){
                            scoutReturnTacixs.push(selectedHandCardix) ;
                            svg.hand.moveToDeckPlayer();
                            used=true;
                        }
                    }else{
                        if(deck===DECK_TROOP){
                            scoutReturnTroopixs.push(selectedHandCardix) ;
                            svg.hand.moveToDeckPlayer();
                            used=true;
                        }
                    }
                    let moves=turn.getMoves();
                    for(let i=0;i<moves.length;i++){
                        let tacEqual=false;
                        if(moves[i].Tac){
                            if(moves[i].Tac.length===scoutReturnTacixs.length){
                                tacEqual=true;
                                for(let j=0;j<moves[i].Tac.length;j++){
                                    if(scoutReturnTacixs[j]!==moves[i].Tac[j]){
                                        tacEqual=false;
                                        break;
                                    }
                                }
                            }
                        }else{
                            if(scoutReturnTacixs.length===0){
                                tacEqual=true;
                            }
                        }
                        if(tacEqual){
                            let equal=false;
                            if(moves[i].Troop){
                                if(moves[i].Troop.length===scoutReturnTroopixs.length){
                                    equal=true;
                                    for(let j=0;j<moves[i].Troop.length;j++){
                                        if(scoutReturnTroopixs[j]!==moves[i].Troop[j]){
                                            equal=false;
                                            break;
                                        }
                                    }
                                }
                            }else{
                                if(scoutReturnTroopixs.length===0){
                                    equal=true;
                                }
                            }
                            if (equal){
                                turn.isMyTurn=false;
                                ws.actionBuilder(ws.ACT_MOVE).move(0,i).send();
                                scoutReturnTroopixs=[];
                                scoutReturnTacixs=[];
                                break;
                            }
                        }
                    }

                }
            }
            return used;
        };
        function onClickedCone(coneElm,idObj){
            //TODO MAYBE add unSelect move to cone
            let used=false;
            if (turn.isMyTurn&&turn.getState()===TURN_FLAG){
                if (cone.validixs.size===0){
                    let moves=turn.getMoves();
                    let validixs;
                    let max=0;
                    for (let i=0;i<moves.length;i++){
                        if (moves[i].Flags.length>max){
                            max=moves[i].Flags.length;
                            validixs=moves[i].Flags;
                        }
                    }
                    cone.validixs=new Set(validixs);
                }
                let ix=idObj.no-1;
                if (cone.validixs.has(ix)){
                    cone.clickedixs.add(ix);
                    svg.cone.pos(coneElm,2);
                    used=true;
                }
            }
            return used;
        };
        let turnTextArea=document.getElementById("turn-text");
        function move(moveView){
            if (moveView.Move.JsonType!=="MoveSave"){
                turn.isMyTurn=turn.update(moveView);
            }else{
                ends();
                turnTextArea.textContent="Game was saved.";
            }
            showMove(moveView);
        };
        exp.move=move;
        let claimButton =document.getElementById("claim-button");
        let passButton=document.getElementById("pass-button");

        claimButton.onclick=function(){
            if (turn.isMyTurn){
                if(turn.getState()===TURN_FLAG){
                    let moves=turn.getMoves();
                    let equal=false;
                    for(let i=0;i<moves.length;i++){
                        if (moves[i].Flags.length===cone.clickedixs.size){
                            equal=true;
                            for(let j=0;j<moves[i].Flags.length;j++){
                                if (!cone.clickedixs.has(moves[i].Flags[j])){
                                    equal=false;
                                    break;
                                }
                            }
                            if(equal){
                                turn.isMyTurn=false;
                                cone.clickedixs.clear();
                                ws.actionBuilder(ws.ACT_MOVE).move(0,i).send();
                                break;
                            }
                        }
                    }
                    if (!equal){
                        console.log("No legal move was found this should not happen");
                    }
                }
            }
        };
        passButton.onclick=function(){
            if(turn.current.MovesPass){
                turn.isMyTurn=false;
                ws.actionBuilder(ws.ACT_MOVE).move(0,-1).send();
            }
        };
        let stopButton=document.getElementById("stop-button");
        stopButton.onclick=function(){
            if (isPlaying&&!turn.stopped){
                ws.actionBuilder(ws.ACT_SAVE).send();
                if (turn.isMyTurn){
                    turn.isMyTurn=false;
                }
                turn.stopped=true;
                stopButton.disabled=true;
            }
        };
        let giveupButton=document.getElementById("giveup-button");
        giveupButton.onclick=function(){
            if (turn.isMyTurn){
                ws.actionBuilder(ws.ACT_QUIT).send();
                turn.isMyTurn=false;
                giveupButton.disabled=true;
            }
        };
        function ends(){
            turn.current=null;
            turn.isMyTurn=false;
            giveupButton.disabled=true;
            stopButton.disabled=true;
            passButton.disabled=true;
            claimButton.disabled=true;
            ws.actionBuilder(ws.ACT_LIST).send();
        };
        turn.update=function(moveView){
            if (!isPlaying()){
                clear();
                stopButton.disabled=false;
            }else{
                turn.oldState=moveView.State;
            }
            turn.current=moveView;
            let myturn=false;
            let turnText="";
            if (moveView.MyTurn){
                turnText="Your Move: ";
                if(!turn.stopped){
                    myturn=true;
                }
            }else{
                turnText="Opponent Move: ";
            }
            switch (moveView.State){
            case TURN_FLAG:
                turnText=turnText+"Claim Flags";
                break;
            case TURN_HAND:
                turnText=turnText+"Play a Card";
                break;
            case TURN_SCOUTR:
                 turnText=turnText+"Return a Cards to Deck";
                break;
            case TURN_QUIT:
            case TURN_FINISH:
                myturn=false;
                turnText="Game over";
                ends();
                break;
            case TURN_DECK:
            case TURN_SCOUT1:
            case TURN_SCOUT2:
                turnText=turnText+"Draw a Card";
                break;
            }
            turnTextArea.textContent= turnText;
            if  (myturn){
                if (moveView.MovesPass){
                    passButton.disabled=false;
                }else{
                    if (moveView.State!==TURN_FLAG){
                        claimButton.disabled=true;
                    }else{
                        claimButton.disabled=false;
                    }
                }
                giveupButton.disabled=false;
            }else{
                claimButton.disabled=true;
                passButton.disabled=true;
                giveupButton.disabled=true;
            }
            return myturn;
        };
        turn.getState=function(){
            return turn.current.State;
        };
        turn.getHandMove=function(cardix){
            let res=[];
            cardix=""+cardix;
            if (cardix in turn.current.MovesHand){
                res= turn.current.MovesHand[cardix];
            }
            return res;
        };
        turn.getMoves=function(){
            return turn.current.Moves;
        };
        turn.clear=function(){
            turn.current=null;
            turn.isMyTurn=false;
            turn.stopped=false;
            turn.oldState=-1;
            claimButton.disabled=true;
            passButton.disabled=true;
            giveupButton.disabled=true;
            stopButton.disabled=true;
            turnTextArea.textContent="";
        };
        svg.itemClicked=function(elems,centerClick){
            let idObj=svg.id.fromName(elems[0].id);
            let used=false;
            switch (idObj.type){
            case svg.id.ID_Card:
                let clickedCardElm=elems[0];
                let parentIdObj=svg.id.fromName(clickedCardElm.parentNode.id);
                if(parentIdObj.type===svg.id.ID_Hand&&parentIdObj.player){
                    if(svg.hand.selected()){
                        used=true;
                        if(svg.hand.selected().id===clickedCardElm.id){
                            svg.hand.unSelect();
                            if (svg.flag.cardSelected()){
                                svg.flag.cardUnSelect();
                            }
                        }else{
                            svg.hand.move(clickedCardElm,!centerClick);
                        }
                    }else{
                        svg.hand.select(clickedCardElm);
                        used=true;
                    }
                }else{//Flag
                    used=onClickedCard(elems[1],clickedCardElm,null);
                }
                break;
            case svg.id.ID_DeckTroop:
            case svg.id.ID_DeckTac:
                used=onClickedDeck(elems[0],idObj.type);
                break;
            case svg.id.ID_FlagTroop:
            case svg.id.ID_FlagTac:
                used=onClickedCard(elems[0],null,null);
                break;
            case svg.id.ID_Cone:
                used=onClickedCone(elems[0],idObj);
                break;
            case svg.id.ID_DishTroop:
            case svg.id.ID_DishTac:
                used=onClickedCard(null,null,elems[0]);
                break;
            }
            return used;
        };
        return exp;
    }
    function createMsg(document,ws){
        let exp={};

        let msgTextInArea = document.getElementById("message-in-text");
        let msgTextOutArea = document.getElementById("message-out-text");
        let infoTextArea = document.getElementById("info-text");
        let messageSelect=document.getElementById("message-select");
        let msgAudio=document.getElementById("msg-audio");
        let flash={};
        flash.counter=0;
        flash.noFlash=3;
        flash.ivId=0;
        flash.flashColor=document.defaultView.getComputedStyle(msgTextInArea,null).backgroundColor;

        function flashStart(){
            if (flash.counter===0){
                flash.ivId=setInterval(flashAnima,100);
            }
        }
        function flashAnima(){
            flash.counter=flash.counter+1;
            if (flash.counter%2!==0){
                msgTextOutArea.style.backgroundColor=flash.flashColor;
            }else{
                msgTextOutArea.style.backgroundColor="";
            }
            if (flash.noFlash*2===flash.counter){
                clearInterval(flash.ivId);
                flash.ivId=0;
                flash.counter=0;
            }
        }

        function playerIsInSelect(value){
            let options=messageSelect.options;
            let exist=false;
            for(let i=1;i<options.length;i++){
                if (value===options[i].value){
                    exist=true;
                    break;
                }
            }
            return exist;
        }

        function addPlayer(value,name){
            let options=messageSelect.options;
            let opt=document.createElement("OPTION");
            opt.value=value;
            opt.text= name;
            if (!playerIsInSelect(value)){
                messageSelect.add(opt);
                messageSelect.selectedIndex=messageSelect.length-1;
            }
        }
        exp.addPlayer=addPlayer;

        function recieved(m){
            let txt;
            let txtArea=infoTextArea;
            if (m.Name){
                txt=m.Name+" -> "+m.Message+"\n";
                if ((m.Sender) && m.Sender!==-1) {//-1 is System
                    console.log(m);
                    addPlayer(m.Sender.toString(),m.Name);
                        txtArea=msgTextOutArea;
                        txt=m.Name+" -> "+m.Message+"\n";
                }else{
                    txt=m.Message+"\n\n";
                }
            }else{
                txt=m.Message+"\n\n";
            }
            txtArea.value=txt+txtArea.value;
            if (txtArea===msgTextOutArea){
                flashStart();
                msgAudio.play();
            }
        }
        exp.recieved=recieved;
        function playerUpdate(pMap){
            let options=messageSelect.options;
            if (options.length>1){
                let removeIx=[];
                for(let i=1;i<options.length;i++){
                    if (!pMap[options[i].value]){
                        removeIx[removeIx.length]=i;
                    }
                };
                if (removeIx.length>0){
                    for(let i=removeIx.length-1;i>=0;i--){
                        messageSelect.remove(removeIx[i]);
                    }
                }
            }
        }
        exp.playerUpdate=playerUpdate;


        function send(){
            if (messageSelect.value!=="0"){
                let message=msgTextInArea.value;
                let idNo= parseInt(messageSelect.value);
                ws.actionBuilder(ws.ACT_MESS).id(idNo).mess(message).send();
                msgTextInArea.value="";
                let name =messageSelect.options[messageSelect.selectedIndex].text;
                let txt=name+" <- "+message+"\n";
                msgTextOutArea.value=txt+msgTextOutArea.value;
            }
        }
        document.getElementById("send-button").onclick=send;

        return exp;
    }
    function createTable(document,msg,ws,game,cookies){
        const IV_From="From";
        const IV_To="To";
        let exp={};
        let invites={};
        exp.invites={};
        let players={};
        exp.players={};

        function getFieldIx(linkField,headers,useId){
            for (let i=0;i<headers.length; i++){
                let field;
                if (useId){
                    field=headers[i].id;
                }else{
                    field=headers[i].getAttribute("tc-link");
                }
                if (field===linkField){
                    return i;
                }
            }
            return -1;
        }
        //table.getFieldsIx find field indexes assume all fields match
        function getFieldsIx(linkFields,headers,useId){
            let res=[];
            for (let lf of linkFields){
                for (let i=0;i<headers.length; i++){
                    let field;
                    if (useId){
                        field=headers[i].id;
                    }else{
                        field=headers[i].getAttribute("tc-link");
                    }
                    if (lf===field){
                        res.push(i);
                        break;
                    }
                }
            }
            if (res.length!==linkFields.length){
                console.log("Missing field");
            }
            return res;
        }
        invites.recieved=function(invite){
            if(invite.Rejected){
                let name=invites.delete(invite.ReceiverID,true);
                if(name){
                    msg.recieved({Message:name+" declined your invitation."});
                }
            }else{
                invites.replace(invite.InvitorID,invite.InvitorName,false);
                msg.addPlayer(invite.InvitorID.toString(),invite.InvitorName);
            }
        };
        exp.invites.recieved=invites.recieved;

        let iTable=document.getElementById("invites-table");
        let iTableHeaders=iTable.getElementsByTagName("th");
        invites.onRetractButton=function(event){
            let row=event.target.parentNode.parentNode;
            let idix=getFieldIx("ith-id",iTableHeaders,true);
            let idNo=parseInt(row.cells[idix].textContent);
            iTable.deleteRow(row.rowIndex);
            ws.actionBuilder(ws.ACT_INVRETRACT).id(idNo).send();
        };
        invites.onAcceptButton=function(event){
            if(!game.isPlaying()){
                let row=event.target.parentNode.parentNode;
                let idix=getFieldIx("ith-id",iTableHeaders,true);
                let idNo=parseInt(row.cells[idix].textContent);
                ws.actionBuilder(ws.ACT_INVACCEPT).id(idNo).send();
                iTable.deleteRow(row.rowIndex);
            }
        };
        invites.onDeclineButton=function(event){
            let row=event.target.parentNode.parentNode;
            let idix=getFieldIx("ith-id",iTableHeaders,true);
            let idNo=parseInt(row.cells[idix].textContent);
            ws.actionBuilder(ws.ACT_INVDECLINE).id(idNo).send();
            iTable.deleteRow(row.rowIndex);
        };
        invites.clear=function(){
            for (let i=iTable.rows.length-1;i>0;i--){
                iTable.deleteRow(iTable.rows[i].rowIndex);
            }
        };
        exp.invites.clear=invites.clear;

        invites.delete=function(idNo,send){
            let from=IV_From;
            if (send){
                from=IV_To;
            }
            let name="";
            let [idix,nameix,fromix]=getFieldsIx(["ith-id","ith-name","ith-from"],iTableHeaders,true);
            for (let i=1;i<iTable.rows.length;i++){
                let row =iTable.rows[i];
                if(row.cells[idix].textContent===idNo.toString()&&row.cells[fromix].textContent===from){
                    iTable.deleteRow(row.rowIndex);
                    name=row.cells[nameix].textContent;
                    break;
                }
            }
            return name;
        };
        invites.add=function(idNo,name,send){
            let newRow=iTable.insertRow(-1);// -1 is add
            for (let i=0;i<iTableHeaders.length; i++){
                let fieldId=iTableHeaders[i].id;
                let cell=newRow.insertCell(-1);// -1 is add
                let newTxtNode;
                switch (fieldId){
                case "ith-id":
                    newTxtNode=document.createTextNode(idNo);
                    cell.appendChild(newTxtNode);
                    break;
                case "ith-from":
                    if (send){
                        newTxtNode=document.createTextNode(IV_To);
                    }else{
                        newTxtNode=document.createTextNode(IV_From);
                    }
                    cell.appendChild(newTxtNode);
                    break;
                case "ith-name":
                    newTxtNode=document.createTextNode(name);
                    cell.appendChild(newTxtNode);
                    break;
                case "ith-retract":
                    if (send){
                        let btn = document.createElement("BUTTON");
                        btn.onclick=invites.onRetractButton;
                        newTxtNode=document.createTextNode("Retract");
                        btn.appendChild(newTxtNode);
                        cell.appendChild(btn);
                    }
                    break;
                case "ith-accept":
                    if(!send){ 
                        let btn = document.createElement("BUTTON");
                        btn.onclick=invites.onAcceptButton;
                        btn.appendChild(document.createTextNode("Accept"));
                        cell.appendChild(btn);
                    }
                    break;
                case "ith-decline":
                    if(!send){
                        let btn = document.createElement("BUTTON");
                        btn.onclick=invites.onDeclineButton;
                        btn.appendChild(document.createTextNode("Decline"));
                        cell.appendChild(btn);
                    }
                    break;
                }//select
            }//for
        };
        invites.contain=function(idNo,send){
            let ix=0;
            let [idix,fromix]=getFieldsIx(["ith-id","ith-from"],iTableHeaders,true);
            let from=IV_From;
            if (send){
                from=IV_To;
            }
            for (let i=1;i<iTable.rows.length;i++){
                let row =iTable.rows[i];
                if(row.cells[idix].textContent===idNo.toString()&&row.cells[fromix].textContent===from){
                    ix=i;
                    break;
                }
            }
            return ix;
        };
        invites.replace=function(idNo,name,send){
            invites.delete(idNo,send);
            invites.add(idNo,name,send);
        };

        let pTable=document.getElementById("players-table");
        let pTbodyEmpty=pTable.getElementsByTagName("tbody")[0].cloneNode(true);
        let pTableHeaders=pTable.getElementsByTagName("thead")[0].getElementsByTagName("th");
        document.getElementById("update-button").onclick=function(){
            ws.actionBuilder(ws.ACT_LIST).send();
        };
        players.clear=function(){
            pMap=new Map([]);
            players.update(pMap);
        };
        exp.players.clear=players.clear;
        players.update=function(pMap){
            msg.playerUpdate(pMap);
            if(iTable.rows.length>1){
                let idix=getFieldIx("ith-id",iTableHeaders,true);
                for (let i=iTable.rows.length-1;i>0;i--){
                    let idNo=iTable.rows[i].cells[idix].textContent;
                    if(!pMap[idNo]){
                        iTable.deleteRow(i);
                    }
                }
            }
            pTable.removeChild(pTable.getElementsByTagName("tbody")[0]);
            let newBody=pTbodyEmpty.cloneNode(true);
            pTable.appendChild(newBody);
            let players=[];
            for(let k of Object.keys(pMap)){
                players.push(pMap[k]);
            }
            players.sort(function(a,b){
                return a.Name.localeCompare(b.Name);
            });
            for(let p of players){
                let newRow=newBody.insertRow(-1);// -1 is add
                for (let i=0;i<pTableHeaders.length; i++){
                    let field=pTableHeaders[i].getAttribute("tc-link");
                    if (field){
                        let cell=newRow.insertCell(-1);// -1 is add
                        let newTxtNode=document.createTextNode(p[field]);
                        cell.appendChild(newTxtNode);
                    }else{
                        if (p.Name!==cookies.name){
                            if (pTableHeaders[i].id==="pt-inv-butt"&&!p.OppName&&!game.isPlaying()){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                btn.onclick=function(event){
                                    if (!game.isPlaying()){
                                        let cells=event.target.parentNode.parentNode.cells;
                                        let [idix,nameix]=getFieldsIx(["ID","Name"],pTableHeaders);
                                        let idNo=parseInt(cells[idix].textContent);
                                        let name=cells[nameix].textContent;
                                        let send=true;
                                        if (invites.contain(idNo,send)===0){//0 is header so we do
                                            invites.add(idNo,name,send);    //not use -1
                                            ws.actionBuilder(ws.ACT_INVITE).id(idNo).send();
                                            msg.addPlayer(idNo.toString(),name);
                                        }
                                    }else{
                                        ws.actionBuilder(ws.ACT_LIST).send();
                                    }
                                };
                                let newTxtNode=document.createTextNode("I");
                                btn.appendChild(newTxtNode);
                                cell.appendChild(btn);
                            }else if(pTableHeaders[i].id==="pt-watch-butt"&&p.OppName&&p.OppName!==cookies.name){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                let newTxtNode=document.createTextNode("W");
                                btn.appendChild(newTxtNode);
                                cell.appendChild(btn);//TODO implement watch
                            }else if(pTableHeaders[i].id==="pt-msg-butt"){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                btn.onclick=function(event){
                                    let cells=event.target.parentNode.parentNode.cells;
                                    let [idix,nameix]=getFieldsIx(["ID","Name"],pTableHeaders);
                                    msg.addPlayer(cells[idix].textContent,cells[nameix].textContent);
                                };
                                let newTxtNode=document.createTextNode("M");
                                btn.appendChild(newTxtNode);
                                cell.appendChild(btn);
                            }else{
                                newRow.insertCell(-1);
                            }
                        }else{
                            newRow.insertCell(-1);
                        }
                    }
                }
            }
        };
        exp.players.update=players.update;
        return exp;
    }
    function createWs(){
        let exp={};
        let protocol="ws";
        if (location.protocol==="https:"){
            protocol="wss";
        }
        let path=protocol+"://"+location.host+"/in/gamews";
        let conn=new WebSocket(path);

        exp.ACT_MESS       = 1;
	      exp.ACT_INVITE     = 2;
	      exp.ACT_INVACCEPT  = 3;
	      exp.ACT_INVDECLINE = 4;
        exp.ACT_INVRETRACT = 5;
	      exp.ACT_MOVE       = 6;
	      exp.ACT_QUIT       = 7;
	      exp.ACT_WATCH      = 8;
	      exp.ACT_WATCHSTOP  = 9;
	      exp.ACT_LIST       = 10;
        exp.ACT_SAVE       = 11;

        function actionBuilder(aType){
            let res={ActType:aType};
            res.id=function(idNo){
                res.ID=idNo;
                return res;
            };
            res.move=function(cardix,flagix){
                this.Move=[cardix,flagix];
                return this;
            };
            res.mess=function(msg){
                this.Mess=msg;
                return this;
            };
            res.send=function(){
                let act=this.build();
                conn.send(JSON.stringify(act));
            };
            function build(){
                let act={ActType:this.ActType};
                if (this.ID){
                    act.ID=this.ID;
                }
                if (this.Move){
                    act.Move=this.Move;
                }
                if (this.Mess){
                    act.Mess=this.Mess;
                }
                return act;
            }
            res.build=build;
            return res;
        }
        exp.actionBuilder=actionBuilder;

        function send(act){
            conn.send(JSON.stringify(act));
        }
        exp.send=send;

        function addListener(table,msg,game){
            conn.onclose=function(event){
                console.log(event.code);
                console.log(event.reason);
                console.log(event.wasClean);
                if(!event.wasClean){
                    console.log("Unclean close of connection.");
                    let txt="No connection to server. you must login again.\n";
                    msg.recieved({Message:txt});
                }
                game.clear();
                table.players.clear();
                exp.unconnected=true;
            };
            conn.onerror=function(event){
                console.log(event.code);
                console.log(event.reason);
                console.log(event.wasClean);
                game.clear();
                table.players.clear();
                exp.unconnected=true;
                let txt="Connection err. You must login again.\n";
                msg.recieved({Message:txt});
            };
            conn.onmessage=function(event){
                const JT_Mess   = 1;
	              const JT_Invite = 2;
	              const JT_Move   = 3;
                const JT_BenchMove = 4;
	              const JT_List   = 5;
                const JT_CloseCon=6;
                const JT_ClearInvites=7;
                //TODO CLEAN up consolelog
                let json=JSON.parse(event.data);
                console.log(json);
                switch (json.JsonType){
                case JT_List:
                    table.players.update(json.Data);
                    break;
                case JT_Mess:
                    msg.recieved(json.Data);
                    break;
                case JT_Invite:
                    table.invites.recieved(json.Data);
                    break;
                case JT_Move:
                    game.move(json.Data);
                    break;
                case JT_CloseCon:
                    msg.recieved({Message:json.Data.Reason});
                    conn.close();
                    game.clear();
                    table.players.clear();
                    exp.unconnected=true;
                    break;
                case JT_ClearInvites:
                    table.invites.clear();
                }
            };
        }
        exp.addListener=addListener;
        return exp;
    }
    window.onload=function(){

        let cookies={};
        cookies.name=getCookies(document)["name"];
        let svg=createSvg(document);
        let ws=createWs();
        let msg=createMsg(document,ws);
        let game=createGame(document,ws,svg,msg);
        let table=createTable(document,msg,ws,game,cookies);
        ws.addListener(table,msg,game);
        window.onbeforeunload = function(e) {
            if (!ws.unconnected){
                let dialogText="";
                dialogText = "You will be log out";
                e.returnValue = dialogText;
                return dialogText;
            }
            return undefined;
        };
        //TODO CLEAN delete test begin
        svg.hand.drawOpp(true);
        svg.hand.drawOpp(false);
        svg.hand.drawOpp(true);
        svg.hand.drawOpp(true);
        svg.hand.drawOpp(false);

        svg.hand.drawPlayer(40);
        svg.hand.unSelect();
        svg.hand.drawPlayer(8);
        svg.hand.unSelect();
        svg.hand.drawPlayer(19);
        svg.hand.move(document.getElementById("card40"),true);
        svg.hand.drawPlayer(51);
        svg.hand.moveToDishPlayer();
        svg.hand.drawPlayer(61);
        svg.hand.moveToDishPlayer();
        svg.hand.drawPlayer(59);
        svg.hand.moveToFlagPlayer(document.getElementById("pF1TroopGroup"));
        svg.hand.drawPlayer(svg.card.TC_Mud);
        svg.hand.moveToFlagPlayer(document.getElementById("pF1TacGroup"));
        svg.hand.moveToDishOpp(58);
        svg.hand.moveToFlagOpp(57,1);
        svg.hand.moveToDeckOpp(true);
        svg.hand.drawPlayer(18);
        svg.hand.moveToFlagPlayer(document.getElementById("pF2TroopGroup"));
        svg.hand.drawOpp(true);
        svg.hand.drawOpp(true);
        svg.hand.drawPlayer(30);
        svg.hand.unSelect();
        svg.hand.drawPlayer(27);
        svg.hand.moveToFlagPlayer(document.getElementById("pF1TroopGroup"));
        svg.hand.select(document.getElementById("card30"));
        svg.hand.drawPlayer(30);
        svg.hand.moveToFlagPlayer(document.getElementById("pF3TroopGroup"));
        svg.hand.drawPlayer(37);
        svg.hand.unSelect();
        svg.hand.drawPlayer(42);
        svg.flag.cardSelect(document.getElementById("card27"));
        svg.flag.cardToFlagPlayer(document.getElementById("pF2TroopGroup"));
        svg.flag.cardToDish(27);
        svg.cone.pos(2,0);
        svg.cone.pos(1,2);
        svg.cone.pos(3,1);
        window.setTimeout(game.clear,2000);
        // delete test end
    }; //onload

 })(batt);
