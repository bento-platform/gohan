#!/bin/bash

# credits: 
# - https://www.biostars.org/p/78929/

FILE=$1

if [[ "$FILE" == "" ]];then
   echo Missing gz file name!
   exit 0
fi

if [[ -f $FILE ]];then
    echo "$FILE exists"
else
    echo "$FILE doesn't exist"
    exit 0
fi

echo
echo Preprocessing :
echo
echo Splitting $FILE into individual VCFs using PERL
echo 
echo Step 1 : creating common and private txt files -
echo This may take a while...

time zcat $FILE | perl -lane '
if (/^#/) { if (/^##/) { print STDERR } else {
 print STDERR join "\t", @F[0..8]; @samples = @F;
} } else {
 print STDERR join "\t", @F[0..8];
 for ($i = 9; $i <= $#F; $i++) {
  if ($F[$i] =~ /^..[1-9]/) {
   print STDOUT join "\t", $samples[$i], $lc, $F[$i];
} } } $lc++' 2> vcfs/_vcf.common.txt | sort -k1,1 -k2,2n > vcfs/_vcf.private.txt

echo Step 2 : converting common and private txt files to individual VCF files -
echo This also may take a while...

mkdir -p vcfs/split
time perl -lane 'BEGIN {
open IN, "vcfs/_vcf.common.txt" or die $!;
chomp(@common = <IN>); foreach (@common) {
 if (/^##/) { $headers .= "$_\n" } else { $headers .= $_; last }
} close IN }
if ($F[0] ne $previousSample) {
 close OUT if $previousSample;
 open OUT, ">vcfs/split/$F[0].vcf";
 print OUT "$headers\t$F[0]";
} $previousSample = $F[0];
print OUT "$common[$F[1]]\t$F[2]";
END { close OUT }' vcfs/_vcf.private.txt

echo Step 3 : compressing individual VCF files -
echo This also may take a while...

time for file in vcfs/split/*vcf; do
  bgzip -f $file;
  # tabix -fp vcf $file.gz
done

# for file in split/*vcf.gz; do
#  gunzip $file
# done

# rm *vcf.gz
# rm *vcf.gz.tbi

# Clean up
mv vcfs/split/*.vcf.gz vcfs/
rmdir vcfs/split/

rm vcfs/_vcf.private.txt
rm vcfs/_vcf.common.txt
