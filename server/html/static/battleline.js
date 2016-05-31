var batt={};
(function(batt){
    var game={};
    var ws={};
    var table={};
    table.players={};
    table.invites={};
    var msg={};
    var invite={};
    var pList={};
    var id={};
    var svg;

    const ACT_MESS       = 1;
	  const ACT_INVITE     = 2;
	  const ACT_INVACCEPT  = 3;
	  const ACT_INVDECLINE = 4;
	  const ACT_MOVE       = 5;
	  const ACT_QUIT       = 6;
	  const ACT_WATCH      = 7;
	  const ACT_WATCHSTOP  = 8;
	  const ACT_LIST       = 9;

    const TROOP_NO = 60;//TODO maybe card infor could be moved to battSvg
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

    function battSvg(){
        const ID_Cone=1;
        const ID_FlagTroop=2;
        const ID_FlagTac=3;
        const ID_Card=4;
        const ID_DeckTac=5;
        const ID_DeckTroop=6;
        const ID_DishTac=7;
        const ID_DishTroop=8;

        const COLORS_Troops=["#00ff00","#ff0000","#af3dff","#ffff00","#007fff","#ff8c00"];
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
            }
            return id;
        };
        svg.fromId=function(id){
            let res={};
            switch (id.charAt(0)){
            case "k":
                res.type=ID_Cone;
                res.no=id.match(/\d/)[0];
                break;
            case "p":
                let dish=true;
                if (id.search(/Dish/)===-1){
                    dish=false;
                }
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
                res.no=id.match(/\d/)[0];
                res.player=true;
                break;
            case "o":
                if (id.search(/Troop/)===-1){
                    res.type=ID_FlagTac;
                }else{
                    res.type=ID_FlagTroop;
                }
                res.no=id.match(/\d/)[0];
                res.player=false;
                break;
            case "c":
                res.type=ID_Card;
                res.no=id.match(/\d{1,2}/)[0];
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
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e+svg.hand.vSpace;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f;
                    }else{
                        newX=cards[i].transform.baseVal.getItem(0).matrix.e;
                        newY=cards[i].transform.baseVal.getItem(0).matrix.f+svg.hand.vSpace;
                    }
                    cards[i].transform.baseVal.getItem(0).setTranslate(newX,newY);
                }
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
            svg.flag.cardToFlag(cardGroup,flagGroup);
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
            svg.hand.moveToFlagPlayer=function(flagGroup){
                let c=svg.hand.selected;
                svg.hand.unSelect();
                svg.card.leave(c,true);
                svg.card.set(flagGroup,c,false);
            };
            svg.hand.moveToFlagOpp=function(cardNo,flagNo){
                let flagGroup;
                if (cardNo>TROOP_NO){
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
                svg.card.leave(cc);
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
                function hitGroup(group){
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
                            if (hitGroup(troopGroup)){
                                res[res.length]=troopGroup;
                                break;
                            }
                            if (hitGroup(tacGroup)){
                                res[res.length]=troopGroup;
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
            let coneR=firstCone.r.baseVal.value+firstCone.style.strokeWidth;
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
            svg.ItemClicked=function(elems){
                let idObj=svg.fromId(elems[0].id);
                switch (idObj.type){
                case ID_Card:
                    let parentId=elems[0].parentNode.id;
                    if(parentId===pHandGroup.id){
                        if(svg.hand.selected){
                            if(svg.hand.selected.id===elems[0].id){
                                svg.hand.unSelect();
                            }else{
                                //TODO
                            }
                        }else{
                            svg.hand.select(elems[0]);
                        }
                    }else{//Flag
                        //TODO
                    }
                    break;
                case ID_DeckTroop:
                    break;
                case ID_DeckTac:
                    break;
                case ID_FlagTroop:
                    break;
                case ID_FlagTac:
                    break;
                case ID_Cone:
                    break;
                case ID_DishTroop:
                    break;
                case ID_DishTac:
                    break;

                }

            };
            battlelineSvg.onclick=function (event){
                let elms=svg.click.hitElems(event);
                if (elms.length>0){
                    svg.ItemClicked(elms);
                }
            };
        };//init
        return svg;
    };

    svg=battSvg();
    table.buttonGetId=function(event){
        let row=event.target.parentNode.parentNode;
        return parseInt(row.cells[0].textContent);
    };
    table.getFieldIx=function(linkField,headers){
        for (let i=0;i<headers.length; i++){
            let field=headers[i].getAttribute("tc-link");
            if (field===linkField){
                return i;
            }
        }
        return -1;
    };
    table.getFieldsIx=function(linkFields,headers){
        let res=[];
        for (let i=0;i<headers.length; i++){
            let field=headers[i].getAttribute("tc-link");
            for (let lf of linkFields){
                if (lf===field){
                    res.push(i);
                    break;
                }
            }
            if (res.length===linkFields.length){
                break;
            }
        }
        return res;
    };
    function actionBuilder(aType){
        let res={ActType:atype};
        res.id=function(id){
            res.Id=id;
            return res;
        };
        res.move=function(x,y){
            this.Move=[x,y];
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
    window.onload=function(){
        id.name=getCookies(document)["name"];
        svg.init(document);
        ws.conn=new WebSocket("ws://game.rezder.com:8181/in/gamews");
        ws.conn.onclose=function(event){
            console.log(event.code);
            console.log(event.reason);
            console.log(event.wasClean);
        };
        ws.conn.onerror=function(event){
            console.log(event.code);
            console.log(event.reason);
            console.log(event.wasClean);
        };
        ws.conn.onmessage=function(event){
            const RES_MESS   = 1;
	          const RES_INVITE = 2;
	          const RES_MOVE   = 3;
	          const RES_LIST   = 4;
            //TODO clean up consolelog
            console.log(event.data);
            let json=JSON.parse(event.data);
            console.log(event.data);
            console.log(json);
            if (json.ResType===RES_LIST){
                table.players.Update(json.List);
            }
        };
        let pTable=document.getElementById("players-table");
        let pTbodyEmpty=pTable.getElementsByTagName("tbody")[0].cloneNode(true);
        let pTableHeaders=pTbodyEmpty.getElementsByTagName("th");
        let messageSelect=document.getElementById("message-select");
        document.getElementById("update-button").onclick=function(){
            let act=actionBuilder(ACT_LIST).build();
            ws.conn.send(JSON.stringify(act));
        };

        table.players.Update=function(pMap){
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
                let newRow=pTable.insertRow(-1);
                for (let i=0;i<pTableHeaders.length; i++){
                    let field=pTableHeaders[i].getAttribute("tc-link");
                    if (field){
                        let cell=newRow.insertCell(-1);
                        let newTxtNode=document.createTextNode(p[field]);
                        cell.appendChild(newTxtNode);
                    }else{
                        if (p.Name!==id.name){
                            if (pTableHeaders[i].id==="pt-inv-butt"){
                                let cell=newRow.insertCell(-1);
                                let btn = document.createElement("BUTTON");
                                let newTxtNode=document.createTextNode("I");
                                 btn.appendChild(newTxtNode);
                                 cell.appendChild(btn);
                            }else if(pTableHeaders[i].id==="pt-watch-butt"){
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
        msg.send=function(){
            
        };
        document.getElementById("send-button").onclick=msg.send;

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
        svg.hand.drawPlayer(27);
        svg.hand.moveToFlagPlayer(document.getElementById("pF1TroopGroup"));
        svg.hand.drawPlayer(37);
        svg.hand.unSelect();
        svg.hand.drawPlayer(42);
        svg.flag.cardSelect(document.getElementById("card27"));
        svg.flag.cardToFlagPlayer(document.getElementById("pF2TroopGroup"));
        svg.flag.cardToDish(27);
        svg.cone.pos(2,0);
        svg.cone.pos(1,2);
        svg.cone.pos(3,1);
        svg. click.zone.coneHit(225,350);
        // delete test end
    }; //onload
})(batt);
