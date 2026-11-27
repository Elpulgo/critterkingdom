<?xml version="1.0" encoding="UTF-8"?>
<tileset version="1.10" tiledversion="1.11.0" name="tileset_outdoor" tilewidth="128" tileheight="128" tilecount="64" columns="8">
 <image source="../../assets/tiles/tileset_outdoor.png" width="1024" height="1024"/>
 <tile id="0" type="grass"/>
 <tile id="1" type="grass_2"/>
 <tile id="2" type="grass_3"/>
 <tile id="3" type="grass_4"/>
 <tile id="4" type="path"/>
 <tile id="5" type="sand"/>
 <tile id="6" type="water_1">
  <animation>
   <frame tileid="6" duration="300"/>
   <frame tileid="7" duration="300"/>
   <frame tileid="8" duration="300"/>
  </animation>
 </tile>
 <tile id="7" type="water_2"/>
 <tile id="8" type="water_3"/>
 <tile id="9" type="grass_path_0001"/>
 <tile id="10" type="grass_path_0010"/>
 <tile id="11" type="grass_path_0011"/>
 <tile id="12" type="grass_path_0100"/>
 <tile id="13" type="grass_path_0101"/>
 <tile id="14" type="grass_path_0110"/>
 <tile id="15" type="grass_path_0111"/>
 <tile id="16" type="grass_path_1000"/>
 <tile id="17" type="grass_path_1001"/>
 <tile id="18" type="grass_path_1010"/>
 <tile id="19" type="grass_path_1011"/>
 <tile id="20" type="grass_path_1100"/>
 <tile id="21" type="grass_path_1101"/>
 <tile id="22" type="grass_path_1110"/>
 <tile id="23" type="grass_water_0001"/>
 <tile id="24" type="grass_water_0010"/>
 <tile id="25" type="grass_water_0011"/>
 <tile id="26" type="grass_water_0100"/>
 <tile id="27" type="grass_water_0101"/>
 <tile id="28" type="grass_water_0110"/>
 <tile id="29" type="grass_water_0111"/>
 <tile id="30" type="grass_water_1000"/>
 <tile id="31" type="grass_water_1001"/>
 <tile id="32" type="grass_water_1010"/>
 <tile id="33" type="grass_water_1011"/>
 <tile id="34" type="grass_water_1100"/>
 <tile id="35" type="grass_water_1101"/>
 <tile id="36" type="grass_water_1110"/>
 <tile id="37" type="sand_water_0001"/>
 <tile id="38" type="sand_water_0010"/>
 <tile id="39" type="sand_water_0011"/>
 <tile id="40" type="sand_water_0100"/>
 <tile id="41" type="sand_water_0101"/>
 <tile id="42" type="sand_water_0110"/>
 <tile id="43" type="sand_water_0111"/>
 <tile id="44" type="sand_water_1000"/>
 <tile id="45" type="sand_water_1001"/>
 <tile id="46" type="sand_water_1010"/>
 <tile id="47" type="sand_water_1011"/>
 <tile id="48" type="sand_water_1100"/>
 <tile id="49" type="sand_water_1101"/>
 <tile id="50" type="sand_water_1110"/>
 <tile id="51" type="cliff_nw"/>
 <tile id="52" type="cliff_n"/>
 <tile id="53" type="cliff_ne"/>
 <tile id="54" type="cliff_w"/>
 <tile id="55" type="cliff_e"/>
 <tile id="56" type="cliff_sw"/>
 <tile id="57" type="cliff_s"/>
 <tile id="58" type="cliff_se"/>
 <tile id="59" type="cliff_face"/>
 <wangsets>
  <wangset name="ground" type="corner" tile="0">
   <wangcolor name="grass" color="#63c74d" tile="0" probability="1"/>
   <wangcolor name="path" color="#e4a672" tile="4" probability="1"/>
   <wangcolor name="water" color="#0099db" tile="6" probability="1"/>
   <wangcolor name="sand" color="#e8b796" tile="5" probability="1"/>
   <wangtile tileid="0" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="1" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="2" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="3" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="4" wangid="0,2,0,2,0,2,0,2"/>
   <wangtile tileid="5" wangid="0,4,0,4,0,4,0,4"/>
   <wangtile tileid="6" wangid="0,3,0,3,0,3,0,3"/>
   <wangtile tileid="9" wangid="0,1,0,2,0,1,0,1"/>
   <wangtile tileid="10" wangid="0,1,0,1,0,2,0,1"/>
   <wangtile tileid="11" wangid="0,1,0,2,0,2,0,1"/>
   <wangtile tileid="12" wangid="0,2,0,1,0,1,0,1"/>
   <wangtile tileid="13" wangid="0,2,0,2,0,1,0,1"/>
   <wangtile tileid="14" wangid="0,2,0,1,0,2,0,1"/>
   <wangtile tileid="15" wangid="0,2,0,2,0,2,0,1"/>
   <wangtile tileid="16" wangid="0,1,0,1,0,1,0,2"/>
   <wangtile tileid="17" wangid="0,1,0,2,0,1,0,2"/>
   <wangtile tileid="18" wangid="0,1,0,1,0,2,0,2"/>
   <wangtile tileid="19" wangid="0,1,0,2,0,2,0,2"/>
   <wangtile tileid="20" wangid="0,2,0,1,0,1,0,2"/>
   <wangtile tileid="21" wangid="0,2,0,2,0,1,0,2"/>
   <wangtile tileid="22" wangid="0,2,0,1,0,2,0,2"/>
   <wangtile tileid="23" wangid="0,1,0,3,0,1,0,1"/>
   <wangtile tileid="24" wangid="0,1,0,1,0,3,0,1"/>
   <wangtile tileid="25" wangid="0,1,0,3,0,3,0,1"/>
   <wangtile tileid="26" wangid="0,3,0,1,0,1,0,1"/>
   <wangtile tileid="27" wangid="0,3,0,3,0,1,0,1"/>
   <wangtile tileid="28" wangid="0,3,0,1,0,3,0,1"/>
   <wangtile tileid="29" wangid="0,3,0,3,0,3,0,1"/>
   <wangtile tileid="30" wangid="0,1,0,1,0,1,0,3"/>
   <wangtile tileid="31" wangid="0,1,0,3,0,1,0,3"/>
   <wangtile tileid="32" wangid="0,1,0,1,0,3,0,3"/>
   <wangtile tileid="33" wangid="0,1,0,3,0,3,0,3"/>
   <wangtile tileid="34" wangid="0,3,0,1,0,1,0,3"/>
   <wangtile tileid="35" wangid="0,3,0,3,0,1,0,3"/>
   <wangtile tileid="36" wangid="0,3,0,1,0,3,0,3"/>
   <wangtile tileid="37" wangid="0,4,0,3,0,4,0,4"/>
   <wangtile tileid="38" wangid="0,4,0,4,0,3,0,4"/>
   <wangtile tileid="39" wangid="0,4,0,3,0,3,0,4"/>
   <wangtile tileid="40" wangid="0,3,0,4,0,4,0,4"/>
   <wangtile tileid="41" wangid="0,3,0,3,0,4,0,4"/>
   <wangtile tileid="42" wangid="0,3,0,4,0,3,0,4"/>
   <wangtile tileid="43" wangid="0,3,0,3,0,3,0,4"/>
   <wangtile tileid="44" wangid="0,4,0,4,0,4,0,3"/>
   <wangtile tileid="45" wangid="0,4,0,3,0,4,0,3"/>
   <wangtile tileid="46" wangid="0,4,0,4,0,3,0,3"/>
   <wangtile tileid="47" wangid="0,4,0,3,0,3,0,3"/>
   <wangtile tileid="48" wangid="0,3,0,4,0,4,0,3"/>
   <wangtile tileid="49" wangid="0,3,0,3,0,4,0,3"/>
   <wangtile tileid="50" wangid="0,3,0,4,0,3,0,3"/>
  </wangset>
 </wangsets>
</tileset>
