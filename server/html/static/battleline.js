var batt={};
(function(batt){
    var game={};
    game.turn={};
    game.turn.current=null;
    game.cone={};
    game.cone.clickedixs=new Set();
    game.cone.validixs=new Set();
    game.scoutReturnTacixs=[];
    game.scoutReturnTroopixs=[];
    var ws={};
    var table={};
    table.players={};
    table.invites={};
    var msg={};
    var id={};
    var svg;

    const ACT_MESS       = 1;
	  const ACT_INVITE     = 2;
	  const ACT_INVACCEPT  = 3;
	  const ACT_INVDECLINE = 4;
    const ACT_INVRETRACT = 5;
	  const ACT_MOVE       = 6;
	  const ACT_QUIT       = 7;
	  const ACT_WATCH      = 8;
	  const ACT_WATCHSTOP  = 9;
	  const ACT_LIST       = 10;

    const TROOP_NO = 60;//TODO maybe card information could be moved to battSvg
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

    const ID_Cone=1;
    const ID_FlagTroop=2;
    const ID_FlagTac=3;
    const ID_Card=4;
    const ID_DeckTac=5;
    const ID_DeckTroop=6;
    const ID_DishTac=7;
    const ID_DishTroop=8;
    const ID_Hand=9;

    function battSvg(){
        const COLORS_Troops=["#00ff00","#ff0000","#af3dff","#ffff00","#007fff","#ffa500"];
        const COLORS_Names=["Green","Red","Purpel","Yellow","Blue","Orange"];
        const COLOR_CardFrame="#ffffff";
        const COLOR_CardFrameSpecial="#000000";
        let svg={};
        svg.card={};
        svg.hand={};
        svg.hand.vSpace=26;
        svg.hand.hSpace=20;
        svg.flag={};
        svg.cone={};
        svg.click={};

        svg.toId=function(type,no,player){
            let pText;
            if(player){
                pText="p";
            }else{
                pText="o";
            }
            let id;
            switch (type){
            case ID_Cone:
            id="k"+no+"Path";
                break;
            case ID_FlagTroop:
                id=pText+"F"+no+"TroopGroup";
                break;
            case ID_FlagTac:
                id=pText+"F"+no+"TacGroup";
                break;
            case ID_Card:
                id="card"+no;
                break;
            case ID_DeckTac:
                id="deckTacGroup";
                break;
            case ID_DeckTroop:
                id="deckTroopGroup";
                break;
            case ID_DishTac:
                id=pText+"DishTacGroup";
                break;
            case ID_DishTroop:
                id=pText+"DishTroopGroup";
                break;
            case ID_Hand:
                id=pText+"Hand";
                break;
            }
            return id;
        };
        svg.fromId=function(id){
            let res={};
            switch (id.charAt(0)){
            case "k":
                res.type=ID_Cone;
                res.no=parseInt(id.match(/\d/)[0]);
                break;
            case "p":
                let dish=false;
                let hand=false;
                if (id.search(/Dish/)>-1){
                    dish=true;
                }else if (id.search(/Hand/)>-1){
                    hand=true;
                }
                if (hand){
                    res.type=ID_Hand;
                }else{
                    if (id.search(/Troop/)===-1){
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
                        res.no=parseInt(id.match(/\d/)[0]);
                    }
                }
                res.player=true;
                break;
            case "o":
                if (id.search(/Troop/)===-1){
                    res.type=ID_FlagTac;
                }else{
                    res.type=ID_FlagTroop;
                }
                res.no=parseInt(id.match(/\d/)[0]);
                res.player=false;
                break;
            case "c":
                res.type=ID_Card;
                res.no=parseInt(id.match(/\d{1,2}/)[0]);
                break;
            case "d":
                if (id.search(/Troop/)===-1){
                    res.type=ID_DeckTac;
                }else{
                    res.type=ID_DeckTroop;
                }
                break;
            }
            return res;
        };
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

        svg.card.count=function(group){
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
        svg.card.stripChildIds =function(group){
            let all = group.getElementsByTagName("*");

            for (let i=0, max=all.length; i < max; i++) {
                if (all[i].id){
                    all[i].removeAttribute("id");
                }
            }
        };

        svg.card.set=function(group,cardGroup,vertical){
            let vspace=0;
            let hspace=0;
            let cardFrame=cardGroup.getElementsByTagName("rect")[0];
            if (vertical){
                vspace=svg.hand.vSpace;
            }else{
                hspace=svg.hand.hSpace;
            }
            let cards,posX,posY;
            if (group.id.charAt(0)==="p"){
                ({n:cards,x:posX,y:posY}=svg.card.count(group));
            }else{
                let gpid="p"+group.id.substring(1,group.id.length);
                ({x:posX,y:posY}=svg.card.count(document.getElementById(gpid)));
                ({n:cards}=svg.card.count(group));
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
        svg.card.leave=function(cardGroup,vertical){
            let group=cardGroup.parentNode;
            let cards=group.getElementsByTagName("g");
            for(let i=cards.length-1;i>=0;i--){
                if (cards[i].id===cardGroup.id){
                    break;
                }else{
                    let newX,newY;
                    if (vertical){
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e-svg.hand.vSpace;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f;
                    }else{
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f-svg.hand.hSpace;
                    }
                    cards[i].transform.baseVal.getItem(0).setTranslate(newX,newY);
                }
            }
        };
        svg.card.clearGroup=function(group){
            let cards=group.getElementsByTagName("g");
            for(let i=cards.length-1;i>=0;i--){
                group.removeChild(cards[i]);
            }
        };
        svg.card.hit=function(group,x,y){
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
        svg.hand.select=function(cardGroup){
            let matrix = document.createElementNS("http://www.w3.org/2000/svg", "svg").createSVGMatrix();
            matrix=matrix.translate(0,-20);
            let item=cardGroup.transform.baseVal.createSVGTransformFromMatrix(matrix);
            cardGroup.transform.baseVal.appendItem(item);
            svg.hand.selected=cardGroup;
        };
        svg.hand.unSelect=function(){
            svg.hand.selected.transform.baseVal.removeItem(1);
            svg.hand.selected=null;
        };

        svg.flag.cardSelect=function(cardGroup){
            cardGroup.getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrameSpecial;
            svg.flag.cardSelected=cardGroup;
        };
        svg.flag.cardUnSelect=function(){
            svg.flag.cardSelected.getElementsByTagName("rect")[0].style.stroke=COLOR_CardFrame;
            svg.flag.cardSelected=null;
        };
        svg.flag.cardToDish=function(cardX){
            let cardGroup;
            if (cardX.parentNode){
                cardGroup=cardX;
            }else{
                cardGroup=document.getElementById(svg.toId(ID_Card,cardX));
            }
            svg.card.leave(cardGroup,false);
            svg.card.moveToDish(cardGroup);
        };
        svg.flag.cardToFlag=function(cardX,flagX,player){
            let cardGroup;
            if (cardX.parentNode){
                cardGroup=cardX;
            }else{
                cardGroup=document.getElementById(svg.toId(ID_Card,cardX));
            }
            let flagGroup;
            if (flagX.parentNode){
                flagGroup=flagX;
            }else{
                if (svg.fromId(cardGroup.id).no>TROOP_NO){
                    flagGroup=document.getElementById(svg.toId(ID_FlagTac,flagX,player));
                }else{
                    flagGroup=document.getElementById(svg.toId(ID_FlagTroop,flagX,player));
                }
            }
            svg.card.leave(cardGroup,false);
            svg.card.set(flagGroup,cardGroup,false);
        };
        svg.flag.cardToFlagPlayer=function(flagGroup){
            let cardGroup=svg.flag.cardSelected;
            svg.flag.cardUnSelect();
            svg.flag.cardToFlag(cardGroup,flagGroup,true);
        };
        svg.cone.pos=function(coneX,pos){
            let coneCircle;
            if (coneX.parentNode){
                coneCircle=coneX;
            }else{
                coneCircle=document.getElementById(svg.toId(ID_Cone,coneX));
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
        svg.cone.clear=function(){
            for (let i =1;i<10;i++){
                svg.cone.pos(i,1);
            }
        };
        svg.itemClicked=function(elems,centerClick){
        };
        svg.card.colorName=function(cardix){
            return COLORS_Names[Math.floor((cardNo-1)/10)];
        };
        svg.init=function (document){
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


            svg.hand.createCard=function(cardNo){
                let cardGroup;
                let text;
                let color;
                if (cardNo>TROOP_NO){
                    switch(cardNo){
                    case TC_Traitor:
                        text="Traitor";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Alexander:
                        text="Alexander";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Darius:
                        text="Darius";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_123:
                        text="123";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_8:
                        text="8";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Deserter:
                        text="Deserter";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Redeploy:
                        text="Redeploy";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Scout:
                        text="Scout";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Mud:
                        text="Mud";
                        cardGroup=tac.cloneNode(true);
                        break;
                    case TC_Fog:
                        text="Fog";
                        cardGroup=tac.cloneNode(true);
                        break;
                    }
                }else{
                    text=""+cardNo%10;
                    if (text==="0"){
                        text="10";
                    }
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

                cardGroup.id=svg.toId(ID_Card,cardNo,false);
                svg.card.stripChildIds(cardGroup);
                return cardGroup;
            };
            svg.hand.createBack=function(troop){
                let cardGroup;
                if (troop){
                    cardGroup=backTroop.cloneNode(true);
                }else{
                    cardGroup=backTac.cloneNode(true);
                }
                cardGroup.removeAttribute("id");
                svg.card.stripChildIds(cardGroup);
                return cardGroup;
            };
            let pHandGroup=document.getElementById("pHandGroup");
            let deckTacTspan=document.getElementById("deckTacTspan");
            let deckTroopTspan=document.getElementById("deckTroopTspan");
            svg.hand.drawPlayer=function(cardNo){
                if (svg.hand.selected){
                    svg.hand.unSelect();
                }
                let cardGroup=svg.hand.createCard(cardNo);
                svg.card.set(pHandGroup,cardGroup,true);
                svg.hand.select(cardGroup);
                if (cardNo>TROOP_NO){
                    deckTacTspan.textContent=""+deckTacTspan.textContent-1;
                }else{
                    deckTroopTspan.textContent=""+deckTroopTspan.textContent-1;
                }
            };
            let oHandGroup=document.getElementById("oHandGroup");
            svg.hand.drawOpp=function(troop){
                let cardGroup=svg.hand.createBack(troop);
                svg.card.set(oHandGroup,cardGroup,true);
                if (troop){
                    deckTroopTspan.textContent=""+deckTroopTspan.textContent-1;
                }else{
                    deckTacTspan.textContent=""+deckTacTspan.textContent-1;
                }
            };
            svg.hand.countPlayer= function(){
                return svg.card.count(pHandGroup);
            };
            svg.hand.move=function(toCard,before){
                let fromCard=svg.hand.selected;
                svg.hand.unSelect();
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
                            move=svg.hand.vSpace;
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
                            move=-svg.hand.vSpace;
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
            svg.hand.moveToDishPlayer=function(){
                let c=svg.hand.selected;
                svg.hand.unSelect();
                svg.card.leave(c,true);
                svg.card.moveToDish(c,true);
            };
            svg.hand.removeOpp=function(cardNo){
                let cards=oHandGroup.getElementsByTagName("g");
                let color;
                if (cardNo>TROOP_NO){
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
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e-svg.hand.vSpace;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f;
                        cards[i].transform.baseVal.getItem(0).setTranslate(newX,newY);
                    }
                }
            };
            svg.hand.moveToDishOpp=function(cardNo){
                svg.hand.removeOpp(cardNo);
                let c=svg.hand.createCard(cardNo);
                svg.card.moveToDish(c,false);
            };
            let oDishTacGroup=document.getElementById("oDishTacGroup");
            let oDishTroopGroup=document.getElementById("oDishTroopGroup");
            svg.card.moveToDish=function(cardGroup,player){
                let v=svg.fromId(cardGroup.id).no;
                if(player===undefined){//work only for flag
                    player=svg.fromId(cardGroup.parentNode.id).player;
                }
                let dishGroup;
                if (player){
                    if(v>TROOP_NO){
                        dishGroup=pDishTacGroup;
                    }else{
                        dishGroup=pDishTroopGroup;
                    }
                }else{
                    if(v>TROOP_NO){
                        dishGroup=oDishTacGroup;
                    }else{
                        dishGroup=oDishTroopGroup;
                    }
                }
                svg.card.set(dishGroup,cardGroup,false);
            };
            svg.card.clear=function(){
                svg.card.clearGroup(oDishTroopGroup);
                svg.card.clearGroup(oDishTacGroup);
                svg.card.clearGroup(pDishTroopGroup);
                svg.card.clearGroup(pDishTacGroup);
                for(let i=1;i<10;i++){
                    let id= svg.toId(ID_FlagTroop,i,true);
                    svg.card.clearGroup(document.getElementById(id));
                    id= svg.toId(ID_FlagTroop,i,false);
                    svg.card.clearGroup(document.getElementById(id));
                    id= svg.toId(ID_FlagTac,i,true);
                    svg.card.clearGroup(document.getElementById(id));
                    id= svg.toId(ID_FlagTac,i,false);
                    svg.card.clearGroup(document.getElementById(id));
                }
                svg.card.clearGroup(pHandGroup);
                svg.card.clearGroup(oHandGroup);
                deckTacTspan.textContent=""+TAC_NO;
                deckTroopTspan.textContent=""+TROOP_NO;
            };
            svg.hand.moveToFlagPlayer=function(flagX){
                let c=svg.hand.selected;
                let cardix=svg.fromId(c.id).no;
                let flagGroup;
                if (flagX.parentNode){
                    let flagIdObj=svg.fromId(flagX.id);
                    if (cardix===TC_Mud||cardix===TC_Fog){
                        if(flagIdObj.type===ID_FlagTac){
                            flagGroup=flagX;
                        }else{
                            flagGroup=document.getElementById(svg.toId(ID_FlagTac,flagIdObj.no,true));
                        }
                    }else{
                        if(flagIdObj.type===ID_FlagTroop){
                            flagGroup=flagX;
                        }else{
                            flagGroup=document.getElementById(svg.toId(ID_FlagTroop,flagIdObj.no,true));
                        }
                    }
                }else{
                    if (cardix===TC_Mud||cardix===TC_Fog){
                        flagGroup=document.getElementById(svg.toId(ID_FlagTac,flagX));
                    }else{
                        flagGroup=document.getElementById(svg.toId(ID_FlagTroop,flagX));
                    }
                }
                svg.hand.unSelect();
                svg.card.leave(c,true);
                svg.card.set(flagGroup,c,false);
            };
            svg.hand.moveToFlagOpp=function(cardNo,flagNo){
                let flagGroup;
                if (cardNo===TC_Mud||cardNo===TC_Fog){
                    flagGroup=document.getElementById(svg.toId(ID_FlagTac,flagNo,false));
                }else{
                    flagGroup=document.getElementById(svg.toId(ID_FlagTroop,flagNo,false));
                }
                svg.hand.removeOpp(cardNo);
                let c=svg.hand.createCard(cardNo);
                svg.card.set(flagGroup,c,false);
            };
            svg.hand.removePlayer=function(cardGroup){
                cardGroup.parentNode.removeChild(cardGroup);
            };
            svg.hand.moveToDeckPlayer=function(){
                let cc=svg.hand.selected;
                svg.hand.unSelect();
                if (svg.fromId(cc.id).no>TROOP_NO){
                    deckTacTspan.textContent=""+(parseInt(deckTacTspan.textContent)+1);
                }else{
                    deckTroopTspan.textContent=""+(parseInt(deckTroopTspan.textContent)+1);
                }
                svg.card.leave(cc,true);
                svg.hand.removePlayer(cc,true);
            };
            svg.hand.moveToDeckOpp=function(tac){
                if (tac){
                    svg.hand.removeOpp(TROOP_NO+1);
                    deckTacTspan.textContent=""+(parseInt(deckTacTspan.textContent)+1);
                }else{
                    svg.hand.removeOpp(TROOP_NO);
                    deckTroopTspan.textContent=""+(parseInt(deckTroopTspan.textContent)+1);
                }
            };
            let battlelineSvg=document.getElementById("battlelineSvg");
            svg.click.zone={};
            svg.click.hitElems=function(event){ 
                let m = battlelineSvg.getScreenCTM();
                let point = battlelineSvg.createSVGPoint();
                point.x = event.clientX;
                point.y = event.clientY;
                point = point.matrixTransform(m.inverse());

                let res=svg.click.zone.handHit(point.x,point.y);
                if (res.length===0){
                    res=svg.click.zone.deckHit(point.x,point.y);
                }
                if(res.length===0){
                    res=svg.click.zone.flagHit(point.x,point.y,true);
                }
                if(res.length===0){
                    res=svg.click.zone.flagHit(point.x,point.y,false);
                }
                if(res.length===0){
                    res=svg.click.zone.coneHit(point.x,point.y);
                }
                if(res.length===0){
                    res=svg.click.zone.dishHit(point.x,point.y);
                }
                return res;
            };
            svg.click.deaFlagSet=new Set();
            svg.click.deactivateFlag=function(flagNo){
                svg.click.deaFlagSet.add(flagNo);
            };
            svg.click.resetDeactivate=function(){
                svg.click.deaFlagSet.clear();
            };
            let pHandRec=pHandGroup.getElementsByTagName("rect")[0];
            let frameStroke=pHandRec.style.strokeWidth/2;
            let pHandArea=new Area(pHandRec.x.baseVal.value+frameStroke,
                                   pHandRec.x.baseVal.value-frameStroke+pHandRec.width.baseVal.value,
                                   pHandRec.y.baseVal.value+frameStroke-svg.hand.hSpace,
                                   pHandRec.y.baseVal.value-frameStroke+pHandRec.height.baseVal.value
                                  );
            svg.click.zone.handHit=function(x,y){
                let res=[];
                if (pHandArea.hit(x,y)){
                    res=svg.card.hit(pHandGroup,x,y);
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
            svg.click.zone.flagHit=function(x,y,player){
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
                        let troopGroup=document.getElementById(svg.toId(ID_FlagTroop,i,player));
                        res=svg.card.hit(troopGroup,x,y);
                        if (res.length===0){
                            let tacGroup=document.getElementById(svg.toId(ID_FlagTac,i,player));
                            res=svg.card.hit(tacGroup,x,y);
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
            svg.click.zone.deckHit=function(x,y){
                let res=[];
                let troopArea=new Area(backTroopRect.x.baseVal.value,
                                       backTroopRect.x.baseVal.value+backTroopRect.width.baseVal.value,
                                       backTroopRect.y.baseVal.value,
                                       backTroopRect.y.baseVal.value+backTroopRect.width.baseVal.value
                                      );
                if(troopArea.hit(x,y)){
                    res[0]=deckTroopGroup;
                }else{
                    let tacArea=new Area(backTacRect.x.baseVal.value,
                                         backTacRect.x.baseVal.value+backTacRect.width.baseVal.value,
                                         backTacRect.y.baseVal.value,
                                         backTacRect.y.baseVal.value+backTacRect.width.baseVal.value
                                        );
                    if(tacArea.hit(x,y)){
                        res[0]=deckTacGroup;
                    }
                }
                return res;
            };
            let pDishTacRect=pDishTacGroup.getElementsByTagName("rect")[0];
            let pDishTroopRect=pDishTroopGroup.getElementsByTagName("rect")[0];
            svg.click.zone.dishHit=function(x,y){
                let res=[];
                let troopArea=new Area(pDishTroopRect.x.baseVal.value,
                                       pDishTroopRect.x.baseVal.value+pDishTroopRect.width.baseVal.value,
                                       pDishTroopRect.y.baseVal.value,
                                       pDishTroopRect.y.baseVal.value+pDishTroopRect.width.baseVal.value
                                      );
                if(troopArea.hit(x,y)){
                    res[0]=pDishTroopGroup;
                }else{
                    let tacArea=new Area(pDishTacRect.x.baseVal.value,
                                         pDishTacRect.x.baseVal.value+pDishTacRect.width.baseVal.value,
                                         pDishTacRect.y.baseVal.value,
                                         pDishTacRect.y.baseVal.value+pDishTacRect.width.baseVal.value
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
            svg.click.zone.coneHit=function(x,y){
                let res=[];
                if(coneArea.hit(x,y)){
                    for (let i=1;i<10;i++){
                        if(!svg.click.deaFlagSet.has(i)){
                            let coneElm= document.getElementById(svg.toId(ID_Cone,i));
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

            battlelineSvg.onclick=function (event){
                let elms=svg.click.hitElems(event);
                let centerClick=event.which===2;//right click is context menu
                if (elms.length>0){
                    svg.itemClicked(elms,centerClick);
                }
            };
        };//init
        return svg;
    }
    svg=battSvg();
    table.buttonGetId=function(event){
        let row=event.target.parentNode.parentNode;
        return parseInt(row.cells[0].textContent);
    };
    table.getFieldIx=function(linkField,headers,useId){
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
    };
    //table.getFieldsIx find field indexes assume all fields match
    table.getFieldsIx=function(linkFields,headers,useId){
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
    };
    table.invites.recieved=function(invite){
        if(invite.Rejected){
            let name=table.invites.delete(invite.ReceiverId,true);
            if(name){
                msg.recieved({Message:name+" declined your invitation."});
            }
        }else{
            table.invites.replace(invite.InvitorId,invite.InvitorName,false);
        }
    };
    function actionBuilder(aType){
        let res={ActType:aType};
        res.id=function(id){
            res.Id=id;
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
        res.build=function(){
            let act={ActType:this.ActType};
            if (this.Id){
                act.Id=this.Id;
            }
            if (this.Move){
                act.Move=this.Move;
            }
            if (this.Mess){
                act.Mess=this.Mess;
            }
            return act;
        };
        return res;
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
    game.ends=function(){
        game.turn.current=null;
        game.turn.isMyTurn=false;
        let act=actionBuilder(ACT_LIST).build();
        ws.conn.send(JSON.stringify(act));
    };
    game.showMove=function(moveView){
        game.cone.validixs.clear();
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
                    msg.recieved({Message:move.Info}); 
                }
                if (move.Win){
                    msg.recieved({Message:"Congratulation you won the game."});
                }
            }else if (moveView.DeltCardix!==0){
                if (svg.hand.selected){
                    svg.hand.unSelect();
                }
               svg.hand.drawPlayer(moveView.DeltCardix);
            }else if(move.JsonType==="MoveQuit"){
                msg.recieved({Message:"You have lost the game by giving up."});
            }else if(move.JsonType==="MoveRedeployView"){
                if(move.RedeployDishixs.length>0){
                    for(let i=0;i<move.RedeployDishixs.length;i++){
                        svg.flag.cardToDish(move.RedeployDishixs[i]);
                    }
                }
            }
        }else{
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
                for (let i;i<pos.Flags.length;i++){
                    let flag=pos.Flags[i];
                    if (flag.OppFlag){
                        svg.cone.pos(i+1,0);
                    }else if(flag.NeuFlag){
                        svg.cone.pos(i+1,1);
                    }else if(flag.PlayFlag){
                        svg.cone.pos(i+1,2);
                    }
                    for (let j=0;j<flag.OppTroops.length;j++){
                        svg.hand.drawOpp(true);
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
                for (let i;i<pos.OppDishTroops.length;i++){
                    let cardNo=pos.OppDishTroops[i];
                    svg.hand.drawOpp(true);
                    svg.hand.moveToDishOpp(cardNo);
                }
                for (let i;i<pos.OppDishTacs.length;i++){
                    let cardNo=pos.OppDishTacs[i];
                    svg.hand.drawOpp(false);
                    svg.hand.moveToDishOpp(cardNo);
                }
                for (let i;i<pos.DishTroops.length;i++){
                    let cardNo=pos.DishTroops[i];
                    svg.hand.drawPlayer(cardNo);
                    svg.hand.moveToDishPlayer();
                }
                for (let i;i<pos.DishTacs.length;i++){
                    let cardNo=pos.DishTacs[i];
                    svg.hand.drawPlayer(cardNo);
                    svg.hand.moveToDishPlayer();
                }
                for (let i;i<pos.OppHand.length;i++){
                    svg.hand.drawOpp(pos.OppHand[i]);
                }
                for (let i;i<pos.Hand.length;i++){
                    let cardNo=pos.Hand[i];
                    svg.hand.drawOpp(cardNo);
                    svg.hand.unSelect();
                }
                break;
            case "MoveCardFlag":
                svg.hand.moveToFlagOpp(moveView.MoveCardix,move.Flagix+1);
                break;
            case "MoveDeck":
                svg.hand.drawOpp(move.Deck===DECK_TROOP);
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
            case "MoveDeserter":
                svg.hand.moveToDishOpp(moveView.MoveCardix);
                svg.flag.cardToDish(move.Card);
                break;
            case "MoveScoutReturnView":
                svg.hand.moveToDishOpp(TC_Scout);
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
                msg.recieved({Message:"Opponent return "+move.Tac+" tactic cards and "+move.Troop+" troop cards."});
                break;
            case "MoveTraitor":
                svg.hand.moveToDishOpp(moveView.MoveCardix);
                if (move.InFlag>=0){
                    svg.flag.cardToFlag(move.OutCard,move.InFlag+1,false);
                }else{
                    svg.flag.cardToDish(move.OutCard);
                }
                break;
            case "MoveRedeployView":
                svg.hand.moveToDishOpp(moveView.MoveCardix);
                if (move.Move.InFlag>=0){
                    svg.flag.cardToFlag(move.Move.OutCard,move.Move.InFlag+1,false);
                }else{
                    svg.flag.cardToDish(move.Move.OutCard);
                }
                if(move.RedeployDishixs.length>0){
                    for(let i=0;i<move.RedeployDishixs.length;i++){
                        svg.flag.cardToDish(move.RedeployDishixs[i]);
                    }
                }
                break;
            case "MovePass":
                msg.recieved({Message:"Your opponent chose not to play a card."});
                break;
            case "MoveQuit":
                msg.recieved({Message:"Congratulation your opponent have given up."});
                break;
            default:
                console.log("Unsupported move: "+move.JsonType);

            }
        }
    };
    game.cone.clear=function(){
        game.cone.clickedixs.clear();
        game.cone.validixs.clear();
    };
    game.clear=function(){
        game.cone.clear();
        game.turn.clear();
        svg.card.clear();
        svg.cone.clear();
        svg.hand.selected=null;
        svg.flag.cardSelected=null;
        game.scoutReturnTroopixs=[];
        game.scoutReturnTacixs=[];
    };
    game.move=function(moveView){
        if (game.turn.current===null){
            table.invites.clear();
            game.clear();
        }else{
            game.turn.oldState=game.turn.current.State;
        }
        game.turn.current=moveView;
        game.turn.isMyTurn=game.turn.update(game.turn.current);
        game.showMove(moveView);
    };
    game.onClickedCard=function(clickedFlagElm,clickedCardElm,clickedDishElm){
        function moveToFlag(turn,cardix,flagElm){
            let flagIdObj=svg.fromId(flagElm.id);
            let flagNo=flagIdObj.no;
            if (flagIdObj.player){
                let moves=turn.MovesHand[""+cardix];
                if (moves){
                    for(let i=0;i<moves.length;i++){
                        if (moves[i].Flagix===flagNo-1){
                            svg.hand.moveToFlagPlayer(flagElm);
                            game.turn.isMyTurn=false;
                            let act=actionBuilder(ACT_MOVE).move(cardix,i).build();
                            ws.conn.send(JSON.stringify(act));
                            break;
                        }
                    }
                }
            }
        }
        if(game.turn.isMyTurn&&svg.hand.selected!==null&&game.turn.current.State===TURN_HAND){
            let player;
            let clickedFlagix;
            if (clickedFlagElm){
                let flagIdObj=svg.fromId(clickedFlagElm.id);
                player=flagIdObj.player;
                clickedFlagix=flagIdObj.no-1;
            }else{
                player=svg.fromId(clickedDishElm.id).player;
                clickedFlagix=-1;
            }
            let selectedHandCardix=svg.fromId(svg.hand.selected.id).no;
            let turn=game.turn.current;
            if (selectedHandCardix>TROOP_NO){//TAC
                switch (selectedHandCardix){
                case TC_123:
                case TC_8:
                case TC_Fog:
                case TC_Mud:
                case TC_Alexander:
                case TC_Darius:
                    moveToFlag(turn,selectedHandCardix,clickedFlagElm);
                    break;
                case TC_Deserter:
                    if (clickedCardElm&&!player){
                        let clickedCardix=svg.fromId(clickedCardElm.id).no;
                        let dmoves=turn.MovesHand[""+selectedHandCardix];
                        for(let i=0;i<dmoves.length;i++){
                            if (dmoves[i].Card===clickedCardix){
                                svg.hand.moveToDishPlayer();
                                svg.flag.cardToDish(clickedCardix);
                                game.turn.isMyTurn=false;
                                let act=actionBuilder(ACT_MOVE).move(selectedHandCardix,i).build();
                                ws.conn.send(JSON.stringify(act));
                                break;
                            }
                        }
                    }
                    break;
                case TC_Redeploy:
                    if (player){
                        if (!svg.flag.cardSelected){
                            if (clickedCardElm){
                                svg.flag.cardSelect(clickedCardElm);
                            }
                        }else{
                            if (clickedCardElm &&svg.flag.cardSelected.id===clickedCardElm.id){
                                svg.flag.cardUnSelect();
                            }else{
                                let selectedFlagCardix=svg.fromId(svg.flag.cardSelected.id).no;
                                let rmoves=turn.MovesHand[""+selectedHandCardix];
                                for(let i=0;i<rmoves.length;i++){
                                    if (rmoves[i].OutCard===selectedFlagCardix&&rmoves[i].InFlag===clickedFlagix){
                                        svg.hand.moveToDishPlayer();
                                        game.turn.isMyTurn=false;
                                        let act= actionBuilder(ACT_MOVE).move(selectedHandCardix,i).build();
                                        if(clickedFlagix!==-1){
                                            svg.flag.cardToFlagPlayer(clickedFlagElm);
                                        }else{
                                            svg.flag.cardUnSelect();
                                            svg.flag.cardToDish(selectedFlagCardix);
                                        }
                                        ws.conn.send(JSON.stringify(act));
                                        break;
                                    }
                                }
                            }
                        }
                    }
                    break;
                case TC_Traitor:
                    if (player){
                        if (svg.flag.cardSelected){
                            let selectedFlagCardix=svg.fromId(svg.flag.cardSelected.id).no;
                            let tmoves=turn.MovesHand[""+selectedHandCardix];
                            for(let i=0;i<tmoves.length;i++){
                                if (tmoves[i].OutCard===selectedFlagCardix&&tmoves[i].InFlag===clickedFlagix){
                                    svg.hand.moveToDishPlayer();
                                    game.turn.isMyTurn=false;
                                    let act= actionBuilder(ACT_MOVE).move(selectedHandCardix,i).build();
                                    svg.flag.cardToFlagPlayer(clickedFlagElm);
                                    ws.conn.send(JSON.stringify(act));
                                    break;
                                }
                            }
                        }
                    }else{//clicked on opp flag
                        if (!svg.flag.cardSelected){
                            if (clickedCardElm){
                                svg.flag.cardSelect(clickedCardElm);
                            }
                        }else{
                            if (clickedCardElm &&svg.flag.cardSelected.id===clickedCardElm.id){
                                svg.flag.cardUnSelect();
                            }
                        }
                    }
                    break;
                }
            }else{//TROOP
                moveToFlag(turn,selectedHandCardix,clickedFlagElm);
            }
        }
    };
    game.onClickedDeck=function(deckElm,idType){
        if(game.turn.isMyTurn){
            let deck;
            if (idType===ID_DeckTac){
                deck=DECK_TAC;
            }else{
                deck=DECK_TROOP;
            }
            if (game.turn.current.State===TURN_SCOUT1||
                game.turn.current.State===TURN_SCOUT2||
                game.turn.current.State===TURN_DECK){
                let moves=game.turn.current.Moves;
                for(let i=0;i<moves.length;i++){
                    if (moves[i].Deck===deck){
                        game.turn.isMyTurn=false;
                        let act= actionBuilder(ACT_MOVE).move(0,i).build();
                        ws.conn.send(JSON.stringify(act));
                        break;
                    }
                }
            }else if (game.turn.current.State===TURN_HAND && svg.hand.selected &&
                      svg.fromId(svg.hand.selected.id).no===TC_Scout){
                let moves=game.turn.current.MovesHand[""+TC_Scout];
                for(let i=0;i<moves.length;i++){
                    if (moves[i].Deck===deck){
                        svg.hand.moveToDishPlayer();
                        game.turn.isMyTurn=false;
                        let act= actionBuilder(ACT_MOVE).move(TC_Scout,i).build();
                        ws.conn.send(JSON.stringify(act));
                        break;
                    }
                }
            }else if(game.turn.current.State===TURN_SCOUTR && svg.hand.selected){
                let selectedHandCardix=svg.fromId(svg.hand.selected.id).no;
                let handCount;
                if( selectedHandCardix>TROOP_NO){
                    if(deck===DECK_TAC){
                        game.scoutReturnTacixs.push(selectedHandCardix) ;
                        svg.hand.moveToDeckPlayer();
                    }
                }else{
                    if(deck===DECK_TROOP){
                        game.scoutReturnTroopixs.push(selectedHandCardix) ;
                        svg.hand.moveToDeckPlayer();
                    }
                }
                let moves=game.turn.current.Moves;
                for(let i=0;i<moves.length;i++){
                    let tacEqual=false;
                    if(moves[i].Tac){
                        if(moves[i].Tac.length===game.scoutReturnTacixs.length){
                            tacEqual=true;
                            for(let j=0;j<moves[i].Tac.length;j++){
                                if(game.scoutReturnTacixs[j]!==moves[i].Tac[j]){
                                    tacEqual=false;
                                    break;
                                }
                            }
                        }
                    }else{
                        if(game.scoutReturnTacixs.length===0){
                            tacEqual=true;
                        }
                    }
                    if(tacEqual){
                        let equal=false;
                        if(moves[i].Troop){
                            if(moves[i].Troop.length===game.scoutReturnTroopixs.length){
                                equal=true;
                                for(let j=0;j<moves[i].Troop.length;j++){
                                    if(game.scoutReturnTroopixs[j]!==moves[i].Troop[j]){
                                        equal=false;
                                        break;
                                    }
                                }
                            }
                        }else{
                            if(game.scoutReturnTroopixs.length===0){
                                equal=true;
                            }
                        }
                        if (equal){
                            game.turn.isMyTurn=false;
                            let act= actionBuilder(ACT_MOVE).move(0,i).build();
                            ws.conn.send(JSON.stringify(act));
                            game.scoutReturnTroopixs=[];
                            game.scoutReturnTacixs=[];
                            break;
                        }
                    }
                }

            }
        }
    };
    game.onClickedCone= function(coneElm,idObj){
        //TODO maybe add unSelect
        if (game.turn.isMyTurn&&game.turn.current.State===TURN_FLAG){
            if (game.cone.validixs.size===0){
                let moves=game.turn.current.Moves;
                let validixs;
                let max=0;
                for (let i=0;i<moves.length;i++){
                    if (moves[i].Flags.length>max){
                        max=moves[i].Flags.length;
                        validixs=moves[i].Flags;
                    }
                }
                game.cone.validixs=new Set(validixs);
            }
            let ix=idObj.no-1;
            if (game.cone.validixs.has(ix)){
                game.cone.clickedixs.add(ix);
                svg.cone.pos(coneElm,2);
            }
        }
    };
    svg.itemClicked=function(elems,centerClick){
        let idObj=svg.fromId(elems[0].id);
        switch (idObj.type){
        case ID_Card:
            let clickedCardElm=elems[0];
            let parentIdObj=svg.fromId(clickedCardElm.parentNode.id);
            if(parentIdObj.type===ID_Hand&&parentIdObj.player){
                if(svg.hand.selected){
                    if(svg.hand.selected.id===clickedCardElm.id){
                        svg.hand.unSelect();
                        if (svg.flag.cardSelected){
                            svg.flag.cardUnSelect();
                        }
                    }else{
                        svg.hand.move(clickedCardElm,!centerClick);
                    }
                }else{
                    svg.hand.select(clickedCardElm);
                }
            }else{//Flag
                game.onClickedCard(elems[1],clickedCardElm,null);
            }
            break;
        case ID_DeckTroop:
        case ID_DeckTac:
            game.onClickedDeck(elems[0],idObj.type);
            break;
        case ID_FlagTroop:
        case ID_FlagTac:
            game.onClickedCard(elems[0],null,null);
            break;
        case ID_Cone:
            game.onClickedCone(elems[0],idObj);
            break;
        case ID_DishTroop,ID_DishTac:
            game.onClickedCard(null,null,elems[0]);
            break;
        }

    };
    window.onload=function(){
        const IV_From="From";
        const IV_To="To";
        id.name=getCookies(document)["name"];
        svg.init(document);
        ws.conn=new WebSocket("ws://game.rezder.com:8181/in/gamews");
        ws.conn.onclose=function(event){
            console.log(event.code);
            console.log(event.reason);
            console.log(event.wasClean);
            if(!event.wasClean){
                let txt="Lost connection to server.\n"+event.reason;
                msg.recieved({Message:txt});
            }
            game.clear();
        };
        ws.conn.onerror=function(event){
            console.log(event.code);
            console.log(event.reason);
            console.log(event.wasClean);
        };
        ws.conn.onmessage=function(event){
            const JT_Mess   = 1;
	          const JT_Invite = 2;
	          const JT_Move   = 3;
            const JT_BenchMove = 4;
	          const JT_List   = 5;
            const JT_CloseCon=6;
            //TODO clean up consolelog
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
                msg.recieved({Message:json.Data});
                break;
            }

        };
        let iTable=document.getElementById("invites-table");
        let iTableHeaders=iTable.getElementsByTagName("th");

        table.invites.onRetractButton=function(event){
            let row=event.target.parentNode.parentNode;
            let idix=table.getFieldIx("ith-id",iTableHeaders,true);
            let id=parseInt(row.cells[idix].textContent);
            iTable.deleteRow(row.rowIndex);
            let act=actionBuilder(ACT_INVRETRACT).id(id).build();
            ws.conn.send(JSON.stringify(act));
        };
        table.invites.onAcceptButton=function(event){
            if(game.turn.current===null){
                let row=event.target.parentNode.parentNode;
                let idix=table.getFieldIx("ith-id",iTableHeaders,true);
                let id=parseInt(row.cells[idix].textContent);
                let act=actionBuilder(ACT_INVACCEPT).id(id).build();
                iTable.deleteRow(row.rowIndex);
                ws.conn.send(JSON.stringify(act));
            }
        };
        table.invites.onDeclineButton=function(event){
            let row=event.target.parentNode.parentNode;
            let idix=table.getFieldIx("ith-id",iTableHeaders,true);
            let id=parseInt(row.cells[idix].textContent);
            let act=actionBuilder(ACT_INVDECLINE).id(id).build();
            iTable.deleteRow(row.rowIndex);
            ws.conn.send(JSON.stringify(act));
        };
        table.invites.clear=function(){
            for (let i=iTable.rows.length-1;i>0;i--){
                iTable.deleteRow(iTable.rows[i].rowIndex);
            }
        };
        table.invites.delete=function(id,send){
            let from=IV_From;
            if (send){
                from=IV_To;
            }
            let name="";
            let [idix,nameix,fromix]=table.getFieldsIx(["ith-id","ith-name","ith-from"],iTableHeaders,true);
            for (let i=1;i<iTable.rows.length;i++){
                let row =iTable.rows[i];
                if(row.cells[idix].textContent===id.toString()&&row.cells[fromix].textContent===from){
                    iTable.deleteRow(row.rowIndex);
                    name=row.cells[nameix].textContent;
                    break;
                }
            }
            return name;
        };
        table.invites.add=function(id,name,send){
            let newRow=iTable.insertRow(-1);// -1 is add
            for (let i=0;i<iTableHeaders.length; i++){
                let fieldId=iTableHeaders[i].id;
                let cell=newRow.insertCell(-1);// -1 is add
                let newTxtNode;
                switch (fieldId){
                case "ith-id":
                    newTxtNode=document.createTextNode(id);
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
                        btn.onclick=table.invites.onRetractButton;
                        newTxtNode=document.createTextNode("Retract");
                        btn.appendChild(newTxtNode);
                        cell.appendChild(btn);
                    }
                    break;
                case "ith-accept":
                    if(!send){ 
                        let btn = document.createElement("BUTTON");
                        btn.onclick=table.invites.onAcceptButton;
                        btn.appendChild(document.createTextNode("Accept"));
                        cell.appendChild(btn);
                    }
                    break;
                case "ith-decline":
                    if(!send){
                        let btn = document.createElement("BUTTON");
                        btn.onclick=table.invites.onDeclineButton;
                        btn.appendChild(document.createTextNode("Decline"));
                        cell.appendChild(btn);
                    }
                    break;
                }//select
            }//for
        };
        table.invites.contain=function(id,send){
            let ix=0;
            let [idix,fromix]=table.getFieldsIx(["ith-id","ith-from"],iTableHeaders,true);
            let from=IV_From;
            if (send){
                from=IV_To;
            }
            for (let i=1;i<iTable.rows.length;i++){
                let row =iTable.rows[i];
                if(row.cells[idix].textContent===id.toString()&&row.cells[fromix].textContent===from){
                    ix=i;
                    break;
                }
            }
            return ix;
        };
        table.invites.replace=function(id,name,send){
            table.invites.delete(id,send);
            table.invites.add(id,name,send);
        };

        let pTable=document.getElementById("players-table");
        let pTbodyEmpty=pTable.getElementsByTagName("tbody")[0].cloneNode(true);
        let pTableHeaders=pTbodyEmpty.getElementsByTagName("th");
        let messageSelect=document.getElementById("message-select");
        document.getElementById("update-button").onclick=function(){
            let act=actionBuilder(ACT_LIST).build();
            ws.conn.send(JSON.stringify(act));
        };
        table.players.update=function(pMap){
            let options=messageSelect.options;
            if (options.length>2){
                let removeIx=[];
                for(let i=1;i<options.length;i++){
                    if (!pMap[options[i].value]){
                        removeIx[removeIx.length]=options[i];
                    }
                };
                if (removeIx.length>0){
                    for(let i=removeIx.length-1;i>=0;i--){
                        messageSelect.remove(removeIx[i]);
                    }
                }
            }
            if(iTable.rows.length>1){
                let idix=table.getFieldIx("ith-id",iTableHeaders,true);
                for (let i=iTable.rows.length-1;i>0;i--){
                    let id=iTable.rows[i].cells[idix].textContent;
                    if(!pMap[id]){
                        iTable.deleteRow(i);
                    }
                }
            }
            pTable.removeChild(pTable.getElementsByTagName("tbody")[0]);
            pTable.appendChild(pTbodyEmpty.cloneNode(true));
            let players=[];
            for(let k of Object.keys(pMap)){
                players.push(pMap[k]);
            }
            players.sort(function(a,b){
                return a.Name.localeCompare(b.Name);
            });
            for(let p of players){
                let newRow=pTable.insertRow(-1);// -1 is add
                for (let i=0;i<pTableHeaders.length; i++){
                    let field=pTableHeaders[i].getAttribute("tc-link");
                    if (field){
                        let cell=newRow.insertCell(-1);// -1 is add
                        let newTxtNode=document.createTextNode(p[field]);
                        cell.appendChild(newTxtNode);
                    }else{
                        if (p.Name!==id.name){
                            if (pTableHeaders[i].id==="pt-inv-butt"&&!p.OppName&&game.turn.current===null){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                btn.onclick=function(event){
                                    if (game.turn.current===null){
                                        let cells=event.target.parentNode.parentNode.cells;
                                        let [idix,nameix]=table.getFieldsIx(["Id","Name"],pTableHeaders);
                                        let id=parseInt(cells[idix].textContent);
                                        let name=cells[nameix].textContent;
                                        let send=true;
                                        if (table.invites.contain(id,send)===0){//0 is header so we do
                                            table.invites.add(id,name,send);    //not use -1
                                            let act=actionBuilder(ACT_INVITE).id(id).build();
                                            ws.conn.send(JSON.stringify(act));
                                        }
                                    }else{
                                        let act=actionBuilder(ACT_LIST).build();
                                        ws.conn.send(JSON.stringify(act));
                                    }
                                };
                                let newTxtNode=document.createTextNode("I");
                                 btn.appendChild(newTxtNode);
                                 cell.appendChild(btn);
                            }else if(pTableHeaders[i].id==="pt-watch-butt"&&p.OppName){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                let newTxtNode=document.createTextNode("W");
                                btn.appendChild(newTxtNode);
                                cell.appendChild(btn);
                            }else if(pTableHeaders[i].id==="pt-msg-butt"){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                btn.onclick=function(event){
                                    let cells=event.target.parentNode.parentNode.cells;
                                    let [idix,nameix]=table.getFieldsIx(["Id","Name"],pTableHeaders);
                                    let opt=document.createElement("OPTION");
                                    opt.value=cells[idix].textContent;
                                    opt.text= cells[nameix].textContent;
                                    messageSelect.add(opt);
                                    messageSelect.selectedIndex=messageSelect.length-1;
                                };
                                let newTxtNode=document.createTextNode("M");
                                btn.appendChild(newTxtNode);
                                cell.appendChild(btn);
                            }
                        }
                    }
                }
            }
        };
        let msgTextArea = document.getElementById("message-text");
        let infoTextArea = document.getElementById("info-text");
        msg.recieved=function(m){
            let txt;
            if (m.Name){
                 txt=m.Name+" -> "+m.Message+"\n";
            }else{
                txt="Info: "+m.Message+"\n";
            }
            infoTextArea.value=txt+infoTextArea.value;
        };
        msg.send=function(){
            if (messageSelect.value!=="0"){
                let message=msgTextArea.value;
                let id= parseInt(messageSelect.value);
                let act =actionBuilder(ACT_MESS).id(id).mess(message).build();
                ws.conn.send(JSON.stringify(act));
                msgTextArea.value="";
                let name =messageSelect.options[messageSelect.selectedIndex].text;
                let txt=name+" <- "+message+"\n";
                infoTextArea.value=txt+infoTextArea.value;
            }
        };
        document.getElementById("send-button").onclick=msg.send;

        let turnPlayerCell=document.getElementById("turn-player-cell");
        let turnPlayerCellClear=turnPlayerCell.textContent;
        let turnTypeCell=document.getElementById("turn-type-cell");
        let turnTypeCellClear=turnTypeCell.textContent;
        let turnDoneButton=document.getElementById("turn-done-button");
        turnDoneButton.onclick=function(){
            if (game.turn.isMyTurn){
                if(game.turn.current.State===TURN_FLAG){
                    let moves=game.turn.current.Moves;
                    let equal=false;
                    for(let i=0;i<moves.length;i++){
                        if (moves[i].Flags.length===game.cone.clickedixs.size){
                            equal=true;
                            for(let j=0;j<moves[i].Flags.length;j++){
                                if (!game.cone.clickedixs.has(moves[i].Flags[j])){
                                    equal=false;
                                    break;
                                }
                            }
                            if(equal){
                                game.turn.isMyTurn=false;
                                game.cone.clickedixs.clear();
                                let act= actionBuilder(ACT_MOVE).move(0,i).build();
                                ws.conn.send(JSON.stringify(act));
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
        document.getElementById("stop-giveup-button").onclick=function(){
            if (game.turn.current&&!game.turn.gaveup){
                let act=actionBuilder(ACT_QUIT).build();
                ws.conn.send(JSON.stringify(act));
                if (game.turn.isMyTurn){
                    game.turn.isMyTurn=false;
                }
                game.turn.gaveup=true;
            }
        };
        game.turn.update=function(turn){
            let myturn=false;
            if (turn.MyTurn){
                turnPlayerCell.textContent="Your Move";
                if(!game.turn.gaveup){
                    myturn=true;
                }
            }else{
                turnPlayerCell.textContent="Opponent Move";
            }
            let txt;
            switch (turn.State){
            case TURN_FLAG:
                txt="Claim Flags";
                break;
            case TURN_HAND:
                txt="Play a Card";
                break;
            case TURN_SCOUTR:
                txt="Return a Cards to Deck";
                break;
            case TURN_QUIT:
            case TURN_FINISH:
                myturn=false;
                txt="Game is Over";
                game.ends();
                break;
            case TURN_DECK:
            case TURN_SCOUT1:
            case TURN_SCOUT2:
                txt="Draw a Card";
                break;
            }
            turnTypeCell.textContent=txt;
            if  (myturn){
                if (turn.MovesPass){
                    turnDoneButton.disabled=false;
                }else{
                    if (turn.State!==TURN_FLAG){
                        turnDoneButton.disabled=true;
                    }else{
                        turnDoneButton.disabled=false;
                    }
                }
            }else{
                turnDoneButton.disabled=true;
            }
            return myturn;
        };
        game.turn.clear=function(){
            game.turn.current=null;
            game.turn.isMyTurn=false;
            game.turn.gaveup=false;
            game.turn.oldState=-1;
            turnDoneButton.disabled=true;
            turnTypeCell.textContent=turnTypeCellClear;
            turnPlayerCell.textContent=turnPlayerCellClear;
        };

        //TODO delete test begin
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
        svg.hand.drawPlayer(TC_Mud);
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
        svg.click.zone.coneHit(225,350);
        window.setTimeout(game.clear,10000);
        // delete test end
    }; //onload

 })(batt);
